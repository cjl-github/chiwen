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
