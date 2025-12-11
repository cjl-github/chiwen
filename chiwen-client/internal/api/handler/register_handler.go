package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chiwen/client/internal/api/mode"
	"github.com/chiwen/client/internal/service"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Register 负责确保本地 UUID + RSA 密钥 + agent_secret_key
func Register() (string, string, error) {

	// -------------------- 1. 目录 --------------------
	dataDir := viper.GetString("client.data_dir")
	if dataDir == "" {
		dataDir = "~/.ssh"
	}
	dataDir = service.ExpandPath(dataDir)

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", "", fmt.Errorf("mkdir data dir: %w", err)
	}

	uuidPath := filepath.Join(dataDir, viper.GetString("client.uuid_file"))
	privKeyPath := filepath.Join(dataDir, viper.GetString("client.key_file"))
	pubKeyPath := filepath.Join(dataDir, viper.GetString("client.public_key_file"))
	secretPath := filepath.Join(dataDir, viper.GetString("client.agent_secret_file"))

	// -------------------- 2. UUID --------------------
	id, err := service.LoadOrCreateUUID(uuidPath)
	if err != nil {
		return "", "", fmt.Errorf("load uuid: %w", err)
	}
	zap.L().Info("client id ready", zap.String("id", id))

	// -------------------- 3. RSA keys --------------------
	pubPEM, err := service.LoadOrCreateRSAKeys(privKeyPath, pubKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("load rsa: %w", err)
	}
	zap.L().Info("rsa keys ready", zap.String("pub_len", fmt.Sprintf("%d", len(pubPEM))))

	// -------------------- 4. 本地存在 agent_secret → 跳过注册 --------------------
	if b, err := os.ReadFile(secretPath); err == nil && len(bytes.TrimSpace(b)) > 0 {
		agentSecret := string(bytes.TrimSpace(b))
		zap.L().Info("found existing agent_secret_key, skip register", zap.String("path", secretPath))
		return id, agentSecret, nil
	}

	// -------------------- 5. 构造注册请求 --------------------
	nonce, err := service.GenerateNonceBase64(48)
	if err != nil {
		return "", "", err
	}
	timestamp := time.Now().Unix()
	hostname, _ := os.Hostname()

	req := &mode.RegisterRequest{
		Nonce:           nonce,
		Timestamp:       timestamp,
		ID:              id,
		Hostname:        hostname,
		ClientPublicKey: string(pubPEM),
	}

	sign, err := service.SignRegisterRequest(privKeyPath, req)
	if err != nil {
		return "", "", fmt.Errorf("sign register error: %w", err)
	}
	req.Signature = sign

	// -------------------- 6. 发送注册请求 --------------------
	respBody, err := service.SendRegister(req)
	if err != nil {

		// ★ 修改点：Duplicate entry 不返回错误，直接进入 pending
		if bytes.Contains([]byte(err.Error()), []byte("Duplicate entry")) {
			zap.L().Warn("UUID already exists on server, entering pending mode...")

			applyID := id // 服务端以 UUID 做 key

			agentSecret, err := pollForAgentSecret(applyID, privKeyPath,
				10*time.Second, 5*time.Minute)
			if err != nil {
				return "", "", fmt.Errorf("poll agent secret failed: %w", err)
			}

			if err := os.WriteFile(secretPath, []byte(agentSecret), 0o600); err != nil {
				return "", "", fmt.Errorf("write agent secret: %w", err)
			}

			zap.L().Info("agent_secret_key saved (pending approved)",
				zap.String("path", secretPath))

			return id, agentSecret, nil
		}

		return "", "", fmt.Errorf("send register error: %w", err)
	}

	zap.L().Info("register response", zap.String("body", string(respBody)))

	// -------------------- 7. 解析响应 --------------------
	var rr mode.RegisterResponse
	if err := json.Unmarshal(respBody, &rr); err != nil {
		return "", "", fmt.Errorf("json unmarshal: %w", err)
	}

	// 添加调试日志
	zap.L().Debug("parsed register response",
		zap.String("apply_id", rr.ApplyID),
		zap.String("status", rr.Status),
		zap.Bool("has_encrypted_secret", rr.EncryptedSecret != ""),
		zap.Bool("has_agent_secret_key", rr.AgentSecretKey != ""),
		zap.String("message", rr.Message))

	// =============================================================
	// ★ 修改点：支持服务端返回 encrypted_secret 字段（新版协议）
	// =============================================================
	if rr.EncryptedSecret != "" { // ★ 新字段判断
		zap.L().Info("server returned encrypted_secret, decrypting...")

		secretBytes, err := base64.StdEncoding.DecodeString(rr.EncryptedSecret)
		if err != nil {
			zap.L().Error("failed to decode encrypted_secret", zap.Error(err))
			return "", "", fmt.Errorf("decode encrypted_secret: %w", err)
		}

		agentSecret, err := service.DecryptAgentSecret(privKeyPath, secretBytes)
		if err != nil {
			zap.L().Error("failed to decrypt encrypted_secret", zap.Error(err))
			return "", "", fmt.Errorf("decrypt encrypted_secret: %w", err)
		}

		zap.L().Info("successfully decrypted agent_secret, saving to file", zap.String("path", secretPath))
		if err := os.WriteFile(secretPath, []byte(agentSecret), 0o600); err != nil {
			zap.L().Error("failed to write agent_secret to file", zap.Error(err))
			return "", "", fmt.Errorf("write agent secret: %w", err)
		}

		zap.L().Info("agent_secret_key saved from encrypted_secret", zap.String("path", secretPath))
		return id, agentSecret, nil
	}

	// -------------------- 8. pending 状态 → 开始轮询 --------------------
	if rr.Status == "pending" { // 统一检查status
		zap.L().Info("registration pending, waiting for approval", zap.String("apply_id", rr.ApplyID))

		agentSecret, err := pollForAgentSecret(rr.ApplyID, privKeyPath,
			10*time.Second, 5*time.Minute)
		if err != nil {
			return "", "", fmt.Errorf("poll agent secret failed: %w", err)
		}

		if err := os.WriteFile(secretPath, []byte(agentSecret), 0o600); err != nil {
			return "", "", fmt.Errorf("write agent secret: %w", err)
		}

		zap.L().Info("agent_secret_key saved", zap.String("path", secretPath))
		return id, agentSecret, nil
	}

	return "", "", fmt.Errorf("unexpected response: no encrypted secret")
}

// -------------------- 轮询等待函数 --------------------
func pollForAgentSecret(applyID, privKeyPath string,
	interval, timeout time.Duration) (string, error) {

	start := time.Now()

	for {
		rr, err := service.CheckRegisterStatus(applyID)
		if err != nil {
			zap.L().Warn("poll status error, retrying", zap.Error(err))
		} else {
			// 添加调试日志
			zap.L().Debug("poll response received",
				zap.String("status", rr.Status),
				zap.Bool("has_encrypted_secret", rr.EncryptedSecret != ""),
				zap.Bool("has_agent_secret_key", rr.AgentSecretKey != ""),
				zap.String("apply_id", rr.ApplyID))

			// =============================================================
			// ★ 修改点：poll 也支持 encrypted_secret 字段
			// =============================================================
			if rr.EncryptedSecret != "" {
				zap.L().Info("found encrypted_secret, attempting decryption")
				secretBytes, err := base64.StdEncoding.DecodeString(rr.EncryptedSecret)
				if err != nil {
					zap.L().Error("failed to decode encrypted_secret", zap.Error(err))
					return "", fmt.Errorf("decode encrypted_secret: %w", err)
				}
				agentSecret, err := service.DecryptAgentSecret(privKeyPath, secretBytes)
				if err != nil {
					zap.L().Error("failed to decrypt encrypted_secret", zap.Error(err))
					return "", fmt.Errorf("decrypt encrypted_secret: %w", err)
				}
				zap.L().Info("successfully decrypted agent_secret")
				return agentSecret, nil
			} else if rr.Status == "approved" {
				return "", fmt.Errorf("approved but no encrypted secret provided")
			}
		}

		if time.Since(start) > timeout {
			return "", fmt.Errorf("timeout waiting for approval")
		}

		time.Sleep(interval)
	}
}
