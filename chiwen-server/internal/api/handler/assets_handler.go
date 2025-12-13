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

// DeleteAssetHandler 删除资产
func DeleteAssetHandler(c *gin.Context) {
	assetID := c.Param("id")
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Asset ID is required",
		})
		return
	}

	err := mysql.DeleteAsset(assetID)
	if err != nil {
		zap.L().Error("Failed to delete asset", zap.String("asset_id", assetID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete asset",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset deleted successfully",
	})
}

// UpdateAssetLabelsHandler 更新资产备注
func UpdateAssetLabelsHandler(c *gin.Context) {
	assetID := c.Param("id")
	if assetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Asset ID is required",
		})
		return
	}

	var request struct {
		Labels map[string]interface{} `json:"labels"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := mysql.UpdateAssetLabels(assetID, request.Labels)
	if err != nil {
		zap.L().Error("Failed to update asset labels", zap.String("asset_id", assetID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update asset labels",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset labels updated successfully",
	})
}
