package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chiwen/server/internal/data/mysql"
	"go.uber.org/zap"
)

// verifyHeartbeatSignature 使用 agent_secret_key 验证心跳签名
func verifyHeartbeatSignature(secret, signatureB64 string, payload string) error {
	if secret == "" {
		return errors.New("agent secret key is empty")
	}

	// 解码 base64 签名
	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		zap.L().Error("Failed to decode signature base64",
			zap.String("signature", signatureB64),
			zap.Error(err))
		return fmt.Errorf("invalid signature encoding: %v", err)
	}

	// 计算 HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expected := h.Sum(nil)

	// 使用 hmac.Equal 进行安全比较
	if !hmac.Equal(sigBytes, expected) {
		// 记录调试信息，但不暴露完整签名
		zap.L().Warn("Heartbeat signature mismatch",
			zap.String("id_payload", strings.Split(payload, "|")[0]),
			zap.Int("received_len", len(sigBytes)),
			zap.Int("expected_len", len(expected)),
			zap.Int("secret_length", len(secret)))
		return errors.New("invalid heartbeat signature")
	}

	zap.L().Debug("Signature verified successfully")
	return nil
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// buildHeartbeatPayload 构建用于签名的 payload
func buildHeartbeatPayload(id string, timestamp int64, metrics map[string]interface{}) string {
	// 将 metrics 转为 JSON 字符串（紧凑格式）
	data, _ := json.Marshal(metrics)
	return fmt.Sprintf("%s|%d|%s", id, timestamp, string(data))
}

// ProcessHeartbeat 处理心跳请求
func ProcessHeartbeat(id string, timestamp int64, metrics map[string]interface{}, signature string) error {
	zap.L().Info("🫀 Processing heartbeat START",
		zap.String("id", id),
		zap.Int64("timestamp", timestamp),
		zap.String("time", time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")))

	// 1️⃣ 校验时间戳 ±120秒
	now := time.Now().Unix()
	zap.L().Debug("Current time vs timestamp",
		zap.Int64("now", now),
		zap.Int64("timestamp", timestamp),
		zap.Int64("diff", now-timestamp))

	if timestamp <= 0 {
		zap.L().Error("Invalid timestamp", zap.Int64("timestamp", timestamp))
		return errors.New("invalid timestamp: zero or negative")
	}

	diff := abs(now - timestamp)
	if diff > 120 {
		zap.L().Error("Timestamp out of range",
			zap.Int64("diff", diff),
			zap.Int64("max_allowed", 120))
		return fmt.Errorf("timestamp out of range: diff=%d seconds (max allowed: 120)", diff)
	}

	// 2️⃣ 获取 agent_secret_key
	zap.L().Debug("Getting agent secret key", zap.String("id", id))
	secret, err := mysql.GetAgentSecretKeyByID(id)
	if err != nil {
		zap.L().Error("❌ Failed to get agent secret key",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("agent secret key not found: %v", err)
	}

	zap.L().Debug("Got agent secret key",
		zap.String("id", id),
		zap.Int("secret_length", len(secret)))

	if secret == "" {
		zap.L().Error("❌ Agent secret key is empty", zap.String("id", id))
		return errors.New("agent secret key is empty")
	}

	// 3️⃣ 验证签名
	payload := buildHeartbeatPayload(id, timestamp, metrics)

	zap.L().Debug("Built payload for signature verification",
		zap.String("id", id),
		zap.String("payload_first_100", func() string {
			if len(payload) > 100 {
				return payload[:100] + "..."
			}
			return payload
		}()),
		zap.Int("payload_length", len(payload)))

	if err := verifyHeartbeatSignature(secret, signature, payload); err != nil {
		zap.L().Error("❌ Signature verification failed",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("signature verification failed: %v", err)
	}

	zap.L().Info("✅ Signature verified successfully", zap.String("id", id))

	// 4️⃣ 检查资产是否存在
	zap.L().Debug("Checking if asset exists", zap.String("id", id))
	asset, err := mysql.GetAssetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			zap.L().Error("❌ Asset not found",
				zap.String("id", id),
				zap.Error(err))
			return errors.New("asset not found, please register first")
		}
		// 其他数据库错误
		zap.L().Error("❌ Database error when getting asset",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("database error: %v", err)
	}

	if asset == nil {
		zap.L().Error("❌ Asset is nil", zap.String("id", id))
		return errors.New("asset is nil")
	}

	zap.L().Info("Asset found",
		zap.String("id", asset.ID),
		zap.String("status", asset.Status),
		zap.String("hostname", asset.Hostname),
		zap.Time("last_updated", asset.UpdatedAt))

	// 5️⃣ 更新动态信息（JSON格式）
	if err := mysql.UpdateAssetDynamicInfo(id, metrics); err != nil {
		zap.L().Error("Failed to update dynamic info",
			zap.String("id", id),
			zap.Error(err))
		// 继续执行，不立即返回错误
	} else {
		zap.L().Debug("Dynamic info updated", zap.String("id", id))
	}

	// 6️⃣ 更新静态信息（JSON格式）
	if err := mysql.UpdateAssetStaticInfoIfChanged(id, metrics); err != nil {
		zap.L().Warn("Update static info failed (non-critical)",
			zap.String("id", id),
			zap.Error(err))
	} else {
		zap.L().Debug("Static info checked/updated", zap.String("id", id))
	}

	// 7️⃣ 更新时间戳和状态为 online（双重保证）
	if err := mysql.UpdateAssetHeartbeat(id); err != nil {
		zap.L().Error("Failed to update heartbeat timestamp",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	zap.L().Info("✅ Heartbeat processed successfully",
		zap.String("id", id),
		zap.String("hostname", asset.Hostname),
		zap.Time("timestamp", time.Unix(timestamp, 0)))

	return nil
}
