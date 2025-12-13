package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/chiwen/server/internal/data/model"
	"github.com/chiwen/server/internal/data/mysql"
	"go.uber.org/zap"
)

// canonical payload: nonce|timestamp|id|hostname|client_public_key
func buildCanonical(payloadNonce string, timestamp int64, id string, hostname, pubkey string) string {
	return fmt.Sprintf("%s|%d|%s|%s|%s", payloadNonce, timestamp, id, hostname, pubkey)
}

// verifyTimestamp 检查 timestamp 是否在当前时间 +/- seconds 允许范围内
func verifyTimestamp(ts int64, seconds int64) bool {
	now := time.Now().Unix()
	if ts <= 0 {
		return false
	}
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}
	return diff <= seconds
}

// verifyRSASignature 使用客户端公钥验证签名（signature 为 base64 编码）
func verifyRSASignature(pubPEM, signatureB64, payload string) error {
	// decode pubPEM
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return errors.New("invalid public key PEM")
	}
	pubI, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := pubI.(*rsa.PublicKey)
	if !ok {
		return errors.New("not RSA public key")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("invalid signature base64: %w", err)
	}

	h := sha256.Sum256([]byte(payload))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sigBytes); err != nil {
		return err
	}
	return nil
}

// RegisterApply 业务：验证并写入申请
// 参数：nonce, timestamp, id, hostname, clientPubKey, signature
// 返回：error（nil 表示成功）
func RegisterApply(nonce string, timestamp int64, id string, hostname string, clientPubKey string, signature string) error {
	// 1. timestamp 检查（±120 秒）
	if !verifyTimestamp(timestamp, 120) {
		return errors.New("timestamp out of allowed range")
	}

	// 2. 构造 canonical payload 并验签
	payload := buildCanonical(nonce, timestamp, id, hostname, clientPubKey)
	if err := verifyRSASignature(clientPubKey, signature, payload); err != nil {
		return fmt.Errorf("signature verify failed: %w", err)
	}

	// 3. 检查 nonce 重复
	if existing, err := mysql.GetApplyByNonce(nonce); err == nil && existing != nil {
		return errors.New("duplicate nonce")
	}

	// 4. 检查同一客户端是否已有 pending 申请
	if existing, err := mysql.GetApplyByClientID(id); err == nil && existing != nil {
		// 已存在 pending 申请，不插入新记录
		return nil
	}

	// 5. 插入新的申请，使用客户端传过来的 UUID
	apply := &model.AgentRegisterApply{
		ID:           id,
		Nonce:        nonce,
		Hostname:     hostname,
		ApplyStatus:  "pending",
		ClientPubKey: clientPubKey,
		CreatedAt:    time.Now(),
	}

	if err := mysql.CreateAgentApply(apply); err != nil {
		return fmt.Errorf("insert apply failed: %w", err)
	}

	return nil
}

// ApproveApply 管理员审批通过
// 新增参数 agentIP：Agent 连接的远程 IP（从 handler 的 c.RealIP() 获取并传入）
func ApproveApply(clientID string, agentIP string) (encryptedSecret string, err error) {
	apply, err := mysql.GetApplyByClientID(clientID)
	if err != nil || apply == nil {
		return "", errors.New("apply not found or already processed")
	}
	if apply.ApplyStatus == "approved" {
		return "", errors.New("already approved")
	}

	// 生成 secret
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}
	secretStr := base64.StdEncoding.EncodeToString(b)

	// 加密
	block, _ := pem.Decode([]byte(apply.ClientPubKey))
	if block == nil {
		return "", errors.New("invalid client public key PEM")
	}
	pubI, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub, ok := pubI.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not RSA public key")
	}
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(secretStr))
	if err != nil {
		return "", err
	}
	encryptedSecret = base64.StdEncoding.EncodeToString(encryptedBytes)

	// 关键：写入 allowed_users = [5] (允许管理员用户访问)
	allowedUsersJSON := `[5]`

	// === 新增：构建初始 static_info，包含 Agent 连接 IP ===
	var staticInfoJSON string
	if agentIP != "" {
		// 清理端口（RemoteAddr 可能是 ip:port）
		ip := agentIP
		if host, _, splitErr := net.SplitHostPort(agentIP); splitErr == nil {
			ip = host
		}
		// 只添加有效 IP
		if net.ParseIP(ip) != nil {
			// 构建完整的static_info，包含hostname、os和network.ips
			staticInfo := map[string]interface{}{
				"hostname": apply.Hostname,
				"os":       "unknown", // 默认值，客户端会在心跳中更新
				"network": map[string]interface{}{
					"ips": []string{ip},
				},
			}
			if bytes, marshalErr := json.Marshal(staticInfo); marshalErr == nil {
				staticInfoJSON = string(bytes)
			}
		}
	}
	// 如果无法获取有效 IP，留空（原行为）
	if staticInfoJSON == "" {
		staticInfoJSON = "{}"
	}

	// 修改：调用支持 static_info 的 DAO 函数（假设你有 CreateAssetWithStaticInfo 或类似）
	// 如果原 CreateAssetWithAllowedUsers 不支持 static_info，需要扩展它或新建一个函数
	// 这里假设已扩展 mysql.CreateAssetWithStaticAndUsers(apply.ID, apply.Hostname, apply.ClientPubKey, secretStr, staticInfoJSON, allowedUsersJSON)
	err = mysql.CreateAssetWithStaticAndUsers(apply.ID, apply.Hostname, apply.ClientPubKey, secretStr, staticInfoJSON, allowedUsersJSON)
	if err != nil {
		return "", err
	}

	// 更新状态
	if err := mysql.UpdateApplyStatus(apply.ID, "approved"); err != nil {
		return "", err
	}

	zap.L().Info("审批成功，已写入默认权限 allowed_users=[5] 并初始化 IP 到 static_info", zap.String("asset_id", apply.ID), zap.String("initial_ip", agentIP))
	return encryptedSecret, nil
}

// RejectApply 管理员拒绝注册申请
func RejectApply(clientID string) error {
	apply, err := mysql.GetApplyByClientID(clientID)
	if err != nil || apply == nil {
		return errors.New("apply not found")
	}

	if apply.ApplyStatus != "pending" {
		return errors.New("apply is not pending")
	}

	// 更新状态为rejected
	if err := mysql.UpdateApplyStatus(clientID, "rejected"); err != nil {
		return fmt.Errorf("update status failed: %w", err)
	}

	// 删除申请记录（可选，或者保留记录用于审计）
	if err := mysql.DeleteApply(clientID); err != nil {
		zap.L().Warn("Failed to delete rejected apply", zap.String("id", clientID), zap.Error(err))
		// 不返回错误，因为状态已经更新为rejected
	}

	zap.L().Info("申请已拒绝", zap.String("id", clientID))
	return nil
}
