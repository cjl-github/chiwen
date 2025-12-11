package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// DecryptAgentSecret 使用客户端私钥（PEM 文件）对服务端返回的 RSA 加密内容解密（PKCS1v15）
func DecryptAgentSecret(privPath string, encrypted []byte) (string, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return "", fmt.Errorf("invalid private key PEM")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// PKCS8 尝试
		if parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes); err2 == nil {
			if pk, ok := parsed.(*rsa.PrivateKey); ok {
				priv = pk
			} else {
				return "", fmt.Errorf("private key is not RSA")
			}
		} else {
			return "", err
		}
	}

	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encrypted)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
