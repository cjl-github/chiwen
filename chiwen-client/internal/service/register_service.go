package service

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chiwen/client/internal/api/mode"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// ExpandPath 支持 ~ 展开
func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if p[:2] == "~/" || p == "~" {
		home, _ := os.UserHomeDir()
		if p == "~" {
			return home
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

// LoadOrCreateUUID 如果 uuid 文件存在就读取，否则生成并持久化
func LoadOrCreateUUID(path string) (string, error) {
	if b, err := os.ReadFile(path); err == nil {
		id := string(bytes.TrimSpace(b))
		if id != "" {
			return id, nil
		}
	}
	u := uuid.New().String()
	// 确保写入时的父目录存在（调用方通常已经创建）
	if err := os.WriteFile(path, []byte(u), 0o600); err != nil {
		return "", err
	}
	return u, nil
}

// LoadOrCreateRSAKeys 如果密钥存在则读取公钥 PEM，否则生成密钥对并保存私钥为 PEM，保存公钥到 pubPath，返回公钥 PEM 字节
func LoadOrCreateRSAKeys(privPath, pubPath string) ([]byte, error) {
	// 若私钥已存在，读取并确保公钥文件也存在（否则提取并写入）
	if _, err := os.Stat(privPath); err == nil {
		privPEM, err := os.ReadFile(privPath)
		if err != nil {
			return nil, err
		}
		pubPEM, err := extractPublicKeyPEMFromPriv(privPEM)
		if err != nil {
			return nil, err
		}
		// 如果公钥文件不存在或内容不同，写入公钥文件
		if err := ensureFileHasContent(pubPath, pubPEM, 0o600); err != nil {
			return nil, err
		}
		return pubPEM, nil
	}

	// 生成 2048-bit RSA
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 私钥保存为 PEM（PKCS1）
	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER}
	privPEM := pem.EncodeToMemory(privBlock)

	// 保存私钥文件，权限 0600
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		return nil, err
	}

	// 导出公钥 PEM（PKIX, ASN.1 DER）
	pubASN1, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}
	pubPEM := pem.EncodeToMemory(pubBlock)

	// 保存公钥文件
	if err := os.WriteFile(pubPath, pubPEM, 0o600); err != nil {
		return nil, err
	}

	return pubPEM, nil
}

// extractPublicKeyPEMFromPriv 从私钥 PEM 提取公钥 PEM
func extractPublicKeyPEMFromPriv(privPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("invalid private key PEM")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 也可能是 PKCS8 格式
		if parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes); err2 == nil {
			if pk, ok := parsed.(*rsa.PrivateKey); ok {
				pubASN1, err := x509.MarshalPKIXPublicKey(&pk.PublicKey)
				if err != nil {
					return nil, err
				}
				pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}
				return pem.EncodeToMemory(pubBlock), nil
			}
		}
		return nil, err
	}
	pubASN1, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}
	return pem.EncodeToMemory(pubBlock), nil
}

// 确保文件存在且内容符合预期，只在必要时才写入文件
func ensureFileHasContent(path string, content []byte, perm os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		// 存在则读取对比（可选）
		existing, err := os.ReadFile(path)
		if err == nil && bytes.Equal(bytes.TrimSpace(existing), bytes.TrimSpace(content)) {
			return nil
		}
	}
	return os.WriteFile(path, content, perm)
}

// GenerateNonceBase64 随机字节并 Base64 编码
func GenerateNonceBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// signRegisterRequest：对规范化 payload 签名
// 签名策略：对 JSON 的 subset 做 deterministic canonical string（例如 nonce|timestamp|id|hostname|client_public_key）并 sha256+rsaPKCS1v15 签名
func SignRegisterRequest(privPath string, req *mode.RegisterRequest) (string, error) {
	// 读取私钥
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return "", fmt.Errorf("invalid private key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// canonical string：用 | 连接字段，保证定长顺序
	payload := fmt.Sprintf("%s|%d|%s|%s|%s", req.Nonce, req.Timestamp, req.ID, req.Hostname, req.ClientPublicKey)
	h := sha256.Sum256([]byte(payload))

	signBytes, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signBytes), nil
}

// sendRegister POST 注册请求到服务端
func SendRegister(req *mode.RegisterRequest) ([]byte, error) {
	serverHost := viper.GetString("server.host")
	serverPort := viper.GetInt("server.port")
	proto := viper.GetString("server.protocol")
	path := viper.GetString("server.register_path")
	if serverHost == "" {
		serverHost = "localhost"
	}
	if proto == "" {
		proto = "http"
	}
	url := fmt.Sprintf("%s://%s:%d%s", proto, serverHost, serverPort, path)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(viper.GetInt("server.timeout")) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func CheckRegisterStatus(applyID string) (*mode.RegisterResponse, error) {
	// 使用配置拼接完整 URL
	url := fmt.Sprintf("%s://%s:%d%s?apply_id=%s",
		viper.GetString("server.protocol"),
		viper.GetString("server.host"),
		viper.GetInt("server.port"),
		viper.GetString("server.register_status_path"),
		applyID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 添加调试日志
	fmt.Printf("[DEBUG] CheckRegisterStatus response for apply_id=%s: %s\n", applyID, string(body))

	var rr mode.RegisterResponse
	if err := json.Unmarshal(body, &rr); err != nil {
		fmt.Printf("[ERROR] Failed to unmarshal response: %v\n", err)
		return nil, err
	}
	return &rr, nil
}
