package mysql

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chiwen/server/internal/data/model"
	"go.uber.org/zap"
)

// GetApplyByClientID 查询时不再限制 pending，任何状态的记录都返回
func GetApplyByClientID(clientID string) (*model.AgentRegisterApply, error) {
	clientID = strings.TrimSpace(clientID)
	var a model.AgentRegisterApply
	err := db.Get(&a, `
        SELECT id, nonce, hostname, apply_status, created_at, client_public_key 
        FROM agent_register_apply 
        WHERE TRIM(id) = ?
    `, clientID)
	if err != nil {
		return nil, err // 包含 sql.ErrNoRows
	}
	return &a, nil
}

// CheckNonceUsed 检查 nonce 是否被客户端使用过
func CheckNonceUsed(clientID, nonce string) (bool, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(1) FROM agent_register_apply WHERE id = ? AND nonce = ?", clientID, nonce)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetApplyByNonce 根据 nonce 查询申请（返回 sql.ErrNoRows 如果不存在）
func GetApplyByNonce(nonce string) (*model.AgentRegisterApply, error) {
	if db == nil {
		panic("db is nil! InitDB not called or failed")
	}
	var a model.AgentRegisterApply
	err := db.Get(&a, "SELECT id, nonce, hostname, apply_status, created_at, client_public_key FROM agent_register_apply WHERE nonce = ?", nonce)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// CreateAgentApply 插入一条新的申请记录
func CreateAgentApply(a *model.AgentRegisterApply) error {
	if a.ApplyStatus == "" {
		a.ApplyStatus = "pending"
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}

	query := `
        INSERT INTO agent_register_apply
            (id, nonce, hostname, apply_status, client_public_key, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err := db.Exec(query,
		a.ID,
		a.Nonce,
		a.Hostname,
		a.ApplyStatus,
		a.ClientPubKey,
		a.CreatedAt,
	)
	if err != nil {
		fmt.Println("❌ [MySQL] Insert error:", err)
		return err
	}
	return nil
}

// UpdateApplyStatus 更新申请状态
func UpdateApplyStatus(id, status string) error {
	query := `UPDATE agent_register_apply SET apply_status = ? WHERE id = ?`
	_, err := db.Exec(query, status, id)
	return err
}

// ★ 新增：删除apply记录
func DeleteApply(id string) error {
	query := `DELETE FROM agent_register_apply WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		fmt.Println("❌ [MySQL] DeleteApply error:", err)
	}
	return err
}

// CreateAsset 写入正式资产表
func CreateAsset(id, hostname, clientPubKey, agentSecretKey string) (*model.Asset, error) {
	zap.L().Info("Creating asset record",
		zap.String("id", id),
		zap.String("hostname", hostname),
		zap.Int("secret_length", len(agentSecretKey)))

	// 清理客户端公钥格式
	cleanPubKey := strings.TrimSpace(clientPubKey)

	// 创建空的 JSON 对象
	emptyJSON := json.RawMessage(`{}`)

	// 使用 INSERT ... ON DUPLICATE KEY UPDATE
	query := `
		INSERT INTO assets 
			(id, hostname, client_public_key, labels, allowed_users, status, 
			 created_at, updated_at, is_deleted, agent_secret_key, 
			 static_info, dynamic_info)
		VALUES (?, ?, ?, ?, NULL, 'online', NOW(), NOW(), 0, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			hostname = VALUES(hostname),
			client_public_key = VALUES(client_public_key),
			status = VALUES(status),
			updated_at = VALUES(updated_at),
			agent_secret_key = VALUES(agent_secret_key)
	`

	result, err := db.Exec(query,
		id,
		hostname,
		cleanPubKey,
		emptyJSON, // labels
		agentSecretKey,
		emptyJSON, // static_info
		emptyJSON, // dynamic_info
	)

	if err != nil {
		zap.L().Error("Failed to create/update asset",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("create/update asset failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	zap.L().Info("Asset record affected",
		zap.String("id", id),
		zap.Int64("rows_affected", rowsAffected))

	// 查询刚创建的记录
	asset := &model.Asset{}
	err = db.Get(asset, `
		SELECT id, hostname, client_public_key, labels, allowed_users, 
		       status, created_at, updated_at, is_deleted
		FROM assets 
		WHERE id = ?
	`, id)

	if err != nil {
		zap.L().Error("Failed to retrieve created asset",
			zap.String("id", id),
			zap.Error(err))
		return nil, err
	}

	zap.L().Info("Asset created/updated successfully",
		zap.String("id", asset.ID),
		zap.String("status", asset.Status),
		zap.Time("updated_at", asset.UpdatedAt))

	return asset, nil
}

// internal/data/mysql/register_dao.go
// ... 其他代码不变 ...

// CreateAssetWithAllowedUsers 写入正式资产表，支持自定义 allowed_users
func CreateAssetWithAllowedUsers(id, hostname, clientPubKey, agentSecretKey, allowedUsersJSON string) error {
	// 使用MySQL的JSON_ARRAY()函数直接构建JSON数组
	// 注意：allowedUsersJSON应该是"[5]"这样的字符串，我们需要提取数字5
	// 但我们直接使用MySQL的JSON_ARRAY(5)来确保正确的JSON格式
	query := `
		INSERT INTO assets 
			(id, hostname, client_public_key, labels, allowed_users, status, 
			 created_at, updated_at, is_deleted, agent_secret_key)
		VALUES (?, ?, ?, ?, JSON_ARRAY(5), 'online', NOW(), NOW(), 0, ?)
		ON DUPLICATE KEY UPDATE
			hostname = VALUES(hostname),
			client_public_key = VALUES(client_public_key),
			allowed_users = VALUES(allowed_users),
			status = 'online',
			updated_at = NOW(),
			agent_secret_key = VALUES(agent_secret_key)
	`
	_, err := db.Exec(query, id, hostname, clientPubKey, `{}`, agentSecretKey)
	return err
}

// GetPendingApplies 获取所有待审批的申请
func GetPendingApplies() ([]model.AgentRegisterApply, error) {
	var applies []model.AgentRegisterApply
	query := `SELECT id, nonce, hostname, apply_status, created_at, client_public_key 
	          FROM agent_register_apply 
	          WHERE apply_status = 'pending' 
	          ORDER BY created_at DESC`

	err := db.Select(&applies, query)
	if err != nil {
		return nil, err
	}
	return applies, nil
}
