// internal/data/mysql/asset_dao.go   （如果你已经有这个文件就追加，没有就新建

package mysql

import "go.uber.org/zap"

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
