package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/chiwen/server/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TTYHandler TTY相关处理器
type TTYHandler struct {
	ttyService *service.TTYService
}

// NewTTYHandler 创建TTY处理器
func NewTTYHandler() *TTYHandler {
	return &TTYHandler{
		ttyService: service.NewTTYService(),
	}
}

// AuthorizeTTYRequest 授权请求结构
type AuthorizeTTYRequest struct {
	Cols int `form:"cols" json:"cols" binding:"min=20,max=200"`
	Rows int `form:"rows" json:"rows" binding:"min=10,max=60"`
}

// GetAgentSessions 获取Agent需要处理的会话
// GET /api/v1/agent/tty/sessions
func (h *TTYHandler) GetAgentSessions(c *gin.Context) {
	// 1. 获取Agent身份（这里需要根据你的Agent认证系统实现）
	// 临时使用查询参数，实际应该从Agent的签名或Token中获取
	agentID := c.Query("id")
	if agentID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "agent id is required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// 2. 调用业务逻辑
	sessions, err := h.ttyService.GetAgentTTYSessions(agentID)
	if err != nil {
		zap.L().Error("Failed to get agent sessions",
			zap.String("agent_id", agentID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get sessions",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// 3. 返回会话列表
	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// ValidateToken 验证Token（WebSocket连接时使用）
// GET /api/v1/tty/validate?token=xxx
func (h *TTYHandler) ValidateToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token is required",
			"code":  "INVALID_TOKEN",
		})
		return
	}

	session, err := h.ttyService.ValidateTTYToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"code":  "INVALID_TOKEN",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"session": session,
		"message": "token is valid",
	})
}

// AuthorizeTTY 授权TTY访问
// GET /api/v1/assets/{id}/tty/authorize
func (h *TTYHandler) AuthorizeTTY(c *gin.Context) {
	// 1. 获取参数
	assetID := c.Param("id")
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "asset_id is required",
			"code":  "INVALID_PARAMETER",
		})
		return
	}

	// 2. 获取用户ID（这里需要根据你的认证系统实现）
	// 临时使用查询参数，实际应该从JWT或Session中获取
	userID := c.Query("user_id")
	if userID == "" {
		// 开发阶段：使用默认用户ID
		userID = "1"
		zap.L().Warn("Using default user_id for development",
			zap.String("asset_id", assetID))
	}

	// 3. 获取终端尺寸参数
	cols := 80
	rows := 24

	if colsStr := c.Query("cols"); colsStr != "" {
		if c, err := strconv.Atoi(colsStr); err == nil && c >= 20 && c <= 200 {
			cols = c
		}
	}

	if rowsStr := c.Query("rows"); rowsStr != "" {
		if r, err := strconv.Atoi(rowsStr); err == nil && r >= 10 && r <= 60 {
			rows = r
		}
	}

	// 4. 获取浏览器IP
	browserIP := c.ClientIP()

	// 5. 调用业务逻辑
	tokenInfo, err := h.ttyService.AuthorizeTTY(assetID, userID, browserIP, cols, rows)
	if err != nil {
		zap.L().Error("TTY authorization failed",
			zap.String("asset_id", assetID),
			zap.String("user_id", userID),
			zap.String("browser_ip", browserIP),
			zap.Error(err))

		// 根据错误类型返回不同的状态码
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not allowed") ||
			strings.Contains(errMsg, "not in allowed users"):
			c.JSON(http.StatusForbidden, gin.H{
				"error": errMsg,
				"code":  "PERMISSION_DENIED",
			})
		case strings.Contains(errMsg, "rate limit") ||
			strings.Contains(errMsg, "too many"):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": errMsg,
				"code":  "RATE_LIMIT_EXCEEDED",
			})
		case strings.Contains(errMsg, "not online") ||
			strings.Contains(errMsg, "deleted"):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
				"code":  "MACHINE_UNAVAILABLE",
			})
		case strings.Contains(errMsg, "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": errMsg,
				"code":  "ASSET_NOT_FOUND",
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errMsg,
				"code":  "AUTHORIZATION_FAILED",
			})
		}
		return
	}

	// 6. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"token":      tokenInfo.Token,
		"ws_url":     tokenInfo.WsURL,
		"expires_in": tokenInfo.Expires,
		"session": gin.H{
			"cols":    cols,
			"rows":    rows,
			"command": "/bin/bash",
		},
	})
}
