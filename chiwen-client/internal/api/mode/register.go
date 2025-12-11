package mode

// 客户端注册请求
type RegisterRequest struct {
	Nonce           string `json:"nonce" binding:"required"`
	Timestamp       int64  `json:"timestamp" binding:"required"`
	ID              string `json:"id" binding:"required"`
	Hostname        string `json:"hostname" binding:"required"`
	ClientPublicKey string `json:"client_public_key" binding:"required"`
	Signature       string `json:"signature" binding:"required"`
}

type RegisterResponse struct {
	ApplyID         string `json:"apply_id"`         // 申请ID，用于pending状态
	Status          string `json:"status"`           // 状态：pending, approved, rejected
	AgentSecretKey  string `json:"agent_secret_key"` // base64 of RSA-encrypted secret，仅在approved状态时存在
	EncryptedSecret string `json:"encrypted_secret"`
	Message         string `json:"message"` // 可选消息
}

// HeartbeatRequest 客户端发送到服务端的心跳结构
type HeartbeatRequest struct {
	ID        string                 `json:"id" binding:"required"`
	Timestamp int64                  `json:"timestamp" binding:"required"`
	Metrics   map[string]interface{} `json:"metrics" binding:"required"`
	Signature string                 `json:"signature" binding:"required"`
}
