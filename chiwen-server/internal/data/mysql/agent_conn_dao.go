// internal/data/mysql/agent_conn_dao.go
package mysql

// RegisterAgentConnection Agent 认证通过后调用
func RegisterAgentConnection(assetID, wsID, remoteAddr string) error {
	query := `
		INSERT INTO agent_connections (asset_id, ws_id, remote_addr, connected_at, last_ping_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			ws_id = VALUES(ws_id),
			remote_addr = VALUES(remote_addr),
			connected_at = NOW(),
			last_ping_at = NOW()
	`
	_, err := db.Exec(query, assetID, wsID, remoteAddr)
	return err
}

// UpdateAgentPing 更新心跳时间
func UpdateAgentPing(assetID string) error {
	_, err := db.Exec(`
		UPDATE agent_connections 
		SET last_ping_at = NOW() 
		WHERE asset_id = ?`, assetID)
	return err
}

// MarkAssetOfflineIfNoPing 后台任务用：超过60秒没心跳就下线
func MarkAssetOfflineIfNoPing() (int64, error) {
	result, err := db.Exec(`
		UPDATE assets a
		JOIN agent_connections ac ON a.id = ac.asset_id
		SET a.status = 'offline', a.updated_at = NOW()
		WHERE ac.last_ping_at < DATE_SUB(NOW(), INTERVAL 60 SECOND)
		  AND a.status = 'online'
	`)
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// RemoveAgentConnection 连接断开时调用
func RemoveAgentConnection(assetID string) error {
	_, err := db.Exec(`DELETE FROM agent_connections WHERE asset_id = ?`, assetID)
	return err
}
