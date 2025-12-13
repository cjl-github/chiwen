package handler

import (
	"net/http"

	"github.com/chiwen/server/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterStatusHandler 查询客户端注册状态
// GET /api/v1/register/status?apply_id=xxxx
func RegisterStatusHandler(c *gin.Context) {
	applyID := c.Query("apply_id")
	if applyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "apply_id required"})
		return
	}

	status, encryptedSecret, err := service.GetRegisterStatus(applyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           status,
		"encrypted_secret": encryptedSecret, // pending 时为空
	})
}

// ApproveRegisterHandler 管理员审批通过注册申请
// POST /api/v1/register/approve
// Body: { "apply_id": "xxxxxx" }
func ApproveRegisterHandler(c *gin.Context) {
	type Request struct {
		ApplyID string `json:"apply_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// 获取 Agent 的真实 IP（Gin 框架推荐方式）
	agentIP := c.ClientIP() // 自动处理 X-Forwarded-For、X-Real-IP 等，优先级最高

	// 调用 service 层的审批函数（已扩展支持 agentIP 参数）
	encryptedSecret, err := service.ApproveApply(req.ApplyID, agentIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "approve success",
		"encrypted_secret": encryptedSecret,
	})
}

// RejectRegisterHandler 管理员拒绝注册申请
// POST /api/v1/register/reject
// Body: { "apply_id": "xxxxxx" }
func RejectRegisterHandler(c *gin.Context) {
	type Request struct {
		ApplyID string `json:"apply_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := service.RejectApply(req.ApplyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reject success",
	})
}
