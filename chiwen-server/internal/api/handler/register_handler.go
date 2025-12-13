package handler

import (
	"net/http"

	"github.com/chiwen/server/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRequest 客户端请求结构体
type RegisterRequest struct {
	Nonce           string `json:"nonce" binding:"required"`
	Timestamp       int64  `json:"timestamp" binding:"required"`
	ID              string `json:"id" binding:"required"`
	Hostname        string `json:"hostname" binding:"required"`
	ClientPublicKey string `json:"client_public_key" binding:"required"`
	Signature       string `json:"signature" binding:"required"`
}

// RegisterHandler 处理注册请求
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用业务逻辑
	if err := service.RegisterApply(req.Nonce, req.Timestamp, req.ID, req.Hostname, req.ClientPublicKey, req.Signature); err != nil {
		// 业务错误：重复 nonce、验签失败、时间不对等
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 成功：返回 pending 与 apply_id（这里我们用 nonce 作为 apply_id）
	c.JSON(http.StatusOK, gin.H{
		"status":   "pending",
		"apply_id": req.ID,
	})
}

// ApproveRequest 是客户端向服务端发送审批请求时的 JSON 结构体
type ApproveRequest struct {
	ID string `json:"id" binding:"required"` // 客户端申请的唯一 ID（UUID 或其他唯一标识），必填
}

// ApproveHandler 是处理管理员审批申请的 HTTP Handler
func ApproveHandler(c *gin.Context) {
	// 1️⃣ 绑定客户端请求的 JSON 到结构体
	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果请求的 JSON 格式错误或缺少必填字段
		// 返回 HTTP 400 Bad Request，并携带错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取 Agent 的真实 IP（Gin 框架推荐方式）
	agentIP := c.ClientIP() // 自动处理 X-Forwarded-For、X-Real-IP 等，优先级最高

	// 2️⃣ 调用业务层逻辑 ApproveApply
	// 该函数会完成：
	//   - 查询 pending 状态的申请
	//   - 生成 agent_secret_key
	//   - 用客户端公钥加密 secret
	//   - 写入正式资产表 assets
	//   - 更新临时申请状态为 approved
	encryptedSecret, err := service.ApproveApply(req.ID, agentIP)
	if err != nil {
		// 如果审批过程出现任何错误（如申请不存在、数据库写入失败、加密失败等）
		// 返回 HTTP 400 并携带错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3️⃣ 成功审批后返回结果
	// 返回 JSON 给前端：
	//   - status: "approved" 表示审批成功
	//   - encrypted_secret: 用客户端公钥加密后的 agent_secret_key
	c.JSON(http.StatusOK, gin.H{
		"status":           "approved",
		"encrypted_secret": encryptedSecret,
	})
}
