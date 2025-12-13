package handler

import (
	"net/http"

	"github.com/chiwen/server/internal/data/mysql"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PendingAppliesHandler 获取待审批的申请列表
func PendingAppliesHandler(c *gin.Context) {
	applies, err := mysql.GetPendingApplies()
	if err != nil {
		zap.L().Error("Failed to get pending applies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve pending applies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applies": applies,
		"count":   len(applies),
	})
}

// RejectApplyRequest 拒绝申请请求结构体
type RejectApplyRequest struct {
	ID string `json:"id" binding:"required"` // 申请ID
}

// RejectApplyHandler 拒绝申请
func RejectApplyHandler(c *gin.Context) {
	var req RejectApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查申请是否存在且状态为pending
	apply, err := mysql.GetApplyByClientID(req.ID)
	if err != nil || apply == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "apply not found"})
		return
	}

	if apply.ApplyStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "apply is not pending"})
		return
	}

	// 更新状态为rejected
	if err := mysql.UpdateApplyStatus(req.ID, "rejected"); err != nil {
		zap.L().Error("Failed to reject apply", zap.String("id", req.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reject apply",
		})
		return
	}

	// 删除申请记录（可选，或者保留记录用于审计）
	if err := mysql.DeleteApply(req.ID); err != nil {
		zap.L().Warn("Failed to delete rejected apply", zap.String("id", req.ID), zap.Error(err))
		// 不返回错误，因为状态已经更新为rejected
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "rejected",
		"message": "申请已拒绝",
	})
}
