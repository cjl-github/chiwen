package handler

import (
	"net/http"

	"github.com/chiwen/server/internal/data/mysql"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AssetsListHandler 获取资产列表
func AssetsListHandler(c *gin.Context) {
	assets, err := mysql.GetAssetsList()
	if err != nil {
		zap.L().Error("Failed to get assets list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve assets list",
		})
		return
	}

	c.JSON(http.StatusOK, assets)
}
