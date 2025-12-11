// internal/data/model/agent_conn.go
package model

import "time"

type AgentConnection struct {
	ID          int64     `db:"id"`
	AssetID     string    `db:"asset_id"`
	WsID        string    `db:"ws_id"` // 每次连接都不同，用于判断重连
	ConnectedAt time.Time `db:"connected_at"`
	LastPingAt  time.Time `db:"last_ping_at"`
	RemoteAddr  string    `db:"remote_addr"`
}
