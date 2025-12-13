package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/chiwen/server/internal/data/model"
	"go.uber.org/zap"
)

// GetAgentSecretKeyByID 获取 agent_secret_key
func GetAgentSecretKeyByID(id string) (string, error) {
	var secret sql.NullString // 使用 NullString 处理 NULL 值
	err := db.Get(&secret, "SELECT agent_secret_key FROM assets WHERE id = ?", id)
	if err != nil {
		zap.L().Error("GetAgentSecretKeyByID failed",
			zap.String("id", id),
			zap.Error(err))
		return "", err
	}

	if !secret.Valid || secret.String == "" {
		zap.L().Error("Agent secret key is NULL or empty", zap.String("id", id))
		return "", errors.New("agent secret key not found or empty")
	}

	return secret.String, nil
}

// GetAssetByID 获取资产信息
func GetAssetByID(id string) (*model.Asset, error) {
	zap.L().Debug("GetAssetByID called", zap.String("id", id))

	var a model.Asset
	query := `SELECT id, client_public_key, hostname, labels, allowed_users, static_info, dynamic_info, status, 
	                 created_at, updated_at, is_deleted 
	          FROM assets WHERE id = ? AND is_deleted = 0`

	err := db.Get(&a, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			zap.L().Warn("Asset not found in database", zap.String("id", id))
			return nil, fmt.Errorf("asset not found for id: %s", id)
		}
		zap.L().Error("GetAssetByID query failed",
			zap.String("id", id),
			zap.String("query", query),
			zap.Error(err))
		return nil, err
	}

	zap.L().Debug("Asset found",
		zap.String("id", a.ID),
		zap.String("status", a.Status),
		zap.String("hostname", a.Hostname),
		zap.Time("updated_at", a.UpdatedAt),
		zap.Bool("has_allowed_users", a.AllowedUsers.Valid))

	return &a, nil
}

// UpdateAssetDynamicInfo 更新动态信息（JSON格式）
func UpdateAssetDynamicInfo(id string, metrics map[string]interface{}) error {
	dynamicInfo, ok := metrics["dynamic_info"]
	if !ok {
		zap.L().Debug("No dynamic_info in metrics", zap.String("id", id))
		return nil
	}

	// 将 dynamic_info 转换为 JSON
	data, err := json.Marshal(dynamicInfo)
	if err != nil {
		zap.L().Error("Failed to marshal dynamic_info to JSON",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("marshal dynamic_info failed: %w", err)
	}

	// 将 JSON 字符串作为参数传递
	query := `UPDATE assets SET dynamic_info = ?, updated_at = ? WHERE id = ?`
	_, err = db.Exec(query, data, time.Now(), id)
	if err != nil {
		zap.L().Error("UpdateAssetDynamicInfo failed",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	zap.L().Debug("Dynamic info updated",
		zap.String("id", id),
		zap.Int("data_length", len(data)))

	return nil
}

// UpdateAssetStaticInfoIfChanged 更新静态信息（JSON格式）
func UpdateAssetStaticInfoIfChanged(id string, metrics map[string]interface{}) error {
	staticInfo, ok := metrics["static_info"]
	if !ok {
		return nil
	}

	// 将 static_info 转换为 JSON
	data, err := json.Marshal(staticInfo)
	if err != nil {
		zap.L().Warn("Failed to marshal static_info to JSON",
			zap.String("id", id),
			zap.Error(err))
		return nil // 不返回错误，继续执行
	}

	// 使用 JSON 比较来检查是否有变化
	query := `
		UPDATE assets 
		SET static_info = ?, updated_at = ? 
		WHERE id = ? 
		AND (static_info IS NULL OR JSON_CONTAINS(?, static_info) = 0 OR JSON_CONTAINS(static_info, ?) = 0)
	`

	result, err := db.Exec(query, data, time.Now(), id, data, data)
	if err != nil {
		zap.L().Warn("UpdateAssetStaticInfoIfChanged failed (non-critical)",
			zap.String("id", id),
			zap.Error(err))
		return nil // 不返回错误，继续执行
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		zap.L().Debug("Static info updated (changed)",
			zap.String("id", id),
			zap.Int64("rows_affected", rowsAffected))
	}

	return nil
}

// UpdateAssetHeartbeat 心跳成功时更新时间和状态
func UpdateAssetHeartbeat(id string) error {
	query := `UPDATE assets SET status = 'online', updated_at = ? WHERE id = ?`
	result, err := db.Exec(query, time.Now(), id)
	if err != nil {
		zap.L().Error("UpdateAssetHeartbeat failed",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	zap.L().Debug("Heartbeat timestamp updated",
		zap.String("id", id),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}

// GetAssetsList 获取资产列表
func GetAssetsList() ([]model.Asset, error) {
	zap.L().Debug("GetAssetsList called")

	var assets []model.Asset
	query := `SELECT id, client_public_key, hostname, labels, allowed_users, static_info, dynamic_info, status, 
	                 created_at, updated_at, is_deleted 
	          FROM assets WHERE is_deleted = 0 ORDER BY updated_at DESC`

	err := db.Select(&assets, query)
	if err != nil {
		zap.L().Error("GetAssetsList query failed",
			zap.String("query", query),
			zap.Error(err))
		return nil, err
	}

	zap.L().Debug("Assets list retrieved", zap.Int("count", len(assets)))
	return assets, nil
}
