// internal/data/mysql/asset_dao.go   （如果你已经有这个文件就追加，没有就新建

package mysql

import (
	"encoding/json"

	"go.uber.org/zap"
)

// UpdateAssetStatus 快速把机器状态改成 online/offline
func UpdateAssetStatus(assetID, status string) error {
	validStatus := map[string]bool{"online": true, "offline": true, "maintenance": true}
	if !validStatus[status] {
		return nil // 非法状态直接忽略
	}
	_, err := db.Exec(`
		UPDATE assets 
		SET status = ?, updated_at = NOW() 
		WHERE id = ? AND is_deleted = 0`, status, assetID)
	if err != nil {
		zap.L().Error("UpdateAssetStatus failed",
			zap.String("asset_id", assetID), zap.String("status", status), zap.Error(err))
	}
	return err
}

// DeleteAsset 软删除资产（设置 is_deleted = 1）
func DeleteAsset(assetID string) error {
	_, err := db.Exec(`
		UPDATE assets 
		SET is_deleted = 1, updated_at = NOW() 
		WHERE id = ?`, assetID)
	if err != nil {
		zap.L().Error("DeleteAsset failed",
			zap.String("asset_id", assetID), zap.Error(err))
	}
	return err
}

// UpdateAssetLabels 更新资产的labels字段
func UpdateAssetLabels(assetID string, labels map[string]interface{}) error {
	// 将labels转换为JSON字符串
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		zap.L().Error("Failed to marshal labels to JSON",
			zap.String("asset_id", assetID), zap.Error(err))
		return err
	}

	_, err = db.Exec(`
		UPDATE assets 
		SET labels = ?, updated_at = NOW() 
		WHERE id = ? AND is_deleted = 0`, labelsJSON, assetID)
	if err != nil {
		zap.L().Error("UpdateAssetLabels failed",
			zap.String("asset_id", assetID), zap.Error(err))
	}
	return err
}
