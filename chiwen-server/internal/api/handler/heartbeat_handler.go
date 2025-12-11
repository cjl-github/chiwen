package handler

import (
	"net/http"

	"github.com/chiwen/server/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HeartbeatRequest 客户端发送心跳的数据结构
type HeartbeatRequest struct {
	ID        string                 `json:"id" binding:"required"`        // 机器ID
	Timestamp int64                  `json:"timestamp" binding:"required"` // 秒级时间戳
	Metrics   map[string]interface{} `json:"metrics" binding:"required"`   // 动态 + 静态信息
	Signature string                 `json:"signature" binding:"required"` // 用 agent_secret_key 签名
}

// HeartbeatHandler 处理客户端心跳请求
func HeartbeatHandler(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("Heartbeat request bind error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format: " + err.Error(),
			"code":  "INVALID_REQUEST",
		})
		return
	}

	zap.L().Debug("Received heartbeat request",
		zap.String("id", req.ID),
		zap.Int64("timestamp", req.Timestamp),
		zap.Int("metrics_size", len(req.Metrics)))

	// 调用 Service 层处理
	if err := service.ProcessHeartbeat(req.ID, req.Timestamp, req.Metrics, req.Signature); err != nil {
		zap.L().Error("Process heartbeat failed",
			zap.String("id", req.ID),
			zap.Error(err))

		// 返回详细的错误信息
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"id":    req.ID,
			"code":  "HEARTBEAT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "heartbeat processed successfully",
	})
}
