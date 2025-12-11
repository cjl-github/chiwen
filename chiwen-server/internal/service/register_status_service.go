package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"

	"github.com/chiwen/server/internal/data/mysql"
)

// GetRegisterStatus 客户端轮询注册状态
func GetRegisterStatus(applyID string) (status string, encryptedSecret string, err error) {
	// 1. 优先查申请表（任何状态）
	apply, err := mysql.GetApplyByClientID(applyID)
	if err == nil && apply != nil {
		if apply.ApplyStatus == "pending" {
			return "pending", "", nil
		}
		if apply.ApplyStatus == "approved" {
			// 已审批成功 → 重新加密返回 secret
			return encryptSecretForClient(applyID, apply.ClientPubKey)
		}
	}

	// 2. 兜底：查 assets 表（兼容历史数据或异常情况）
	asset, err := mysql.GetAssetByID(applyID)
	if err != nil || asset == nil {
		return "", "", errors.New("apply or asset not found")
	}

	// 从 assets 表拿明文 secret 并重新加密
	return encryptSecretForClient(applyID, asset.ClientPubKey)
}

// 统一加密函数（避免代码重复，也彻底消除 “declared and not used”）
func encryptSecretForClient(id, clientPubKey string) (string, string, error) {
	secret, err := mysql.GetAgentSecretKeyByID(id)
	if err != nil {
		return "", "", err
	}

	block, _ := pem.Decode([]byte(clientPubKey))
	if block == nil {
		return "", "", errors.New("invalid client public key PEM")
	}
	pubI, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", "", err
	}
	pub, ok := pubI.(*rsa.PublicKey)
	if !ok {
		return "", "", errors.New("not RSA public key")
	}

	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(secret))
	if err != nil {
		return "", "", err
	}

	return "approved", base64.StdEncoding.EncodeToString(encryptedBytes), nil
}
