package model

import (
	"time"
)

// TTYSession TTY会话模型
type TTYSession struct {
	ID           string    `db:"id" json:"id"`                       // 会话ID
	AssetID      string    `db:"asset_id" json:"asset_id"`           // 机器ID
	UserID       string    `db:"user_id" json:"user_id"`             // 用户ID
	Token        string    `db:"token" json:"token"`                 // 一次性Token
	Status       string    `db:"status" json:"status"`               // 状态：pending, connected, closed
	Command      string    `db:"command" json:"command"`             // 执行的命令，默认是/bin/bash
	TerminalCols int       `db:"terminal_cols" json:"terminal_cols"` // 终端列数  修改后
	TerminalRows int       `db:"terminal_rows" json:"terminal_rows"` // 终端行数  修改后
	CreatedAt    time.Time `db:"created_at" json:"created_at"`       // 创建时间
	ConnectedAt  time.Time `db:"connected_at" json:"connected_at"`   // 连接时间
	ClosedAt     time.Time `db:"closed_at" json:"closed_at"`         // 关闭时间
	BrowserIP    string    `db:"browser_ip" json:"browser_ip"`       // 浏览器IP
	AgentIP      string    `db:"agent_ip" json:"agent_ip"`           // Agent IP
	RecordFile   string    `db:"record_file" json:"record_file"`     // 录像文件路径
	ErrorMessage string    `db:"error_message" json:"error_message"` // 错误信息
}

// TTYTokenInfo Token信息（不存储，只用于传输）
type TTYTokenInfo struct {
	Token   string `json:"token"`
	WsURL   string `json:"ws_url"`
	Expires int64  `json:"expires_in"` // 过期时间（秒）
}
