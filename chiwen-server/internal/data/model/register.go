package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Assets 表示注册成功的机器资产
type Asset struct {
	ID           string         `db:"id"`                // 机器唯一ID
	ClientPubKey string         `db:"client_public_key"` // Agent秘钥 这个应该客户端的公钥
	Hostname     string         `db:"hostname"`
	Labels       sql.NullString `db:"labels"`        // JSON
	AllowedUsers sql.NullString `db:"allowed_users"` // ★ 新增：允许的用户列表（JSON格式）

	Status    string    `db:"status"` // online/offline/maintenance
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	IsDeleted bool      `db:"is_deleted"`
}

// GetLabelsJSON 安全获取 Labels JSON
func (a *Asset) GetLabelsJSON() (map[string]interface{}, error) {
	if !a.Labels.Valid || a.Labels.String == "" {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(a.Labels.String), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAllowedUsersArray 安全获取 AllowedUsers 数组
func (a *Asset) GetAllowedUsersArray() ([]string, error) {
	if !a.AllowedUsers.Valid || a.AllowedUsers.String == "" {
		return []string{}, nil
	}

	var result []string
	err := json.Unmarshal([]byte(a.AllowedUsers.String), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 临时申请表
type AgentRegisterApply struct {
	ID           string    `db:"id" json:"id"`
	Nonce        string    `db:"nonce" json:"nonce"`
	Hostname     string    `db:"hostname" json:"hostname"`
	ApplyStatus  string    `db:"apply_status" json:"apply_status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	ClientPubKey string    `db:"client_public_key" json:"client_public_key"`
}
