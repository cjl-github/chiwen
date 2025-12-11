package handler

import (
	"net/http"

	"github.com/chiwen/server/internal/data/mysql"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DiagnosticHandler 提供诊断信息
func DiagnosticHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	// 1. 检查 assets 表
	asset, err := mysql.GetAssetByID(id)
	if err != nil {
		zap.L().Warn("Asset not found in diagnostic", zap.String("id", id))
	} else {
		zap.L().Info("Asset found in diagnostic",
			zap.String("id", id),
			zap.String("status", asset.Status))
	}

	// 2. 获取 secret
	secret, err := mysql.GetAgentSecretKeyByID(id)
	if err != nil {
		zap.L().Warn("Secret not found in diagnostic", zap.String("id", id))
	}

	result := gin.H{
		"id": id,
		"assets_table": gin.H{
			"exists": asset != nil,
			"status": func() string {
				if asset != nil {
					return asset.Status
				} else {
					return ""
				}
			}(),
			"updated_at": func() string {
				if asset != nil {
					return asset.UpdatedAt.String()
				} else {
					return ""
				}
			}(),
		},
		"secret": gin.H{
			"exists": secret != "",
			"length": len(secret),
		},
	}

	c.JSON(http.StatusOK, result)
}

// DatabaseStatsHandler 显示数据库统计信息
func DatabaseStatsHandler(c *gin.Context) {
	var assetsCount, ttySessionsCount int
	var latestAssets, latestTTY string

	// 获取 assets 表统计
	_ = mysql.DB().Get(&assetsCount, "SELECT COUNT(*) FROM assets")
	_ = mysql.DB().Get(&latestAssets, "SELECT MAX(updated_at) FROM assets")

	// 获取 tty_sessions 表统计
	_ = mysql.DB().Get(&ttySessionsCount, "SELECT COUNT(*) FROM tty_sessions")
	_ = mysql.DB().Get(&latestTTY, "SELECT MAX(created_at) FROM tty_sessions")

	c.JSON(http.StatusOK, gin.H{
		"assets": gin.H{
			"count":         assetsCount,
			"latest_update": latestAssets,
		},
		"tty_sessions": gin.H{
			"count":          ttySessionsCount,
			"latest_session": latestTTY,
		},
	})
}
