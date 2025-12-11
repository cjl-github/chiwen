package mysql

import (
	"errors"
	"time"

	"github.com/chiwen/server/internal/data/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateTTYSessionStatus 更新会话状态
func UpdateTTYSessionStatus(id, status string) error {
	query := `UPDATE tty_sessions SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, id)
	if err != nil {
		zap.L().Error("Failed to update TTY session status",
			zap.String("session_id", id),
			zap.String("status", status),
			zap.Error(err))
	}
	return err
}

// UpdateTTYSessionConnected 标记会话已连接
func UpdateTTYSessionConnected(id, agentIP string) error {
	query := `UPDATE tty_sessions SET status = 'connected', connected_at = ?, agent_ip = ? WHERE id = ?`
	_, err := db.Exec(query, time.Now(), agentIP, id)
	if err != nil {
		zap.L().Error("Failed to update TTY session connected",
			zap.String("session_id", id),
			zap.Error(err))
	}
	return err
}

// GetActiveSessionsCount 获取机器的活跃会话数
func GetActiveSessionsCount(assetID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM tty_sessions 
		WHERE asset_id = ? AND status IN ('pending', 'connected')
		  AND created_at > DATE_SUB(NOW(), INTERVAL 30 MINUTE)
	`
	err := db.Get(&count, query, assetID)
	return count, err
}

// GetUserActiveSessionsCount 获取用户的活跃会话数
func GetUserActiveSessionsCount(userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM tty_sessions 
		WHERE user_id = ? AND status IN ('pending', 'connected')
		  AND created_at > DATE_SUB(NOW(), INTERVAL 30 MINUTE)
	`
	err := db.Get(&count, query, userID)
	return count, err
}

// CleanupExpiredSessions 清理过期的会话（30分钟前创建的pending会话）
func CleanupExpiredSessions() error {
	query := `
		UPDATE tty_sessions 
		SET status = 'closed', 
		    error_message = 'Session expired',
		    closed_at = NOW()
		WHERE status = 'pending' 
		  AND created_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE)
	`

	result, err := db.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		zap.L().Info("Cleaned up expired TTY sessions",
			zap.Int64("count", rowsAffected))
	}
	return nil
}

// CreateTTYSession 创建TTY会话
func CreateTTYSession(session *model.TTYSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.Status == "" {
		session.Status = "pending"
	}
	if session.Command == "" {
		session.Command = "/bin/bash"
	}
	if session.TerminalCols == 0 {
		session.TerminalCols = 80
	}
	if session.TerminalRows == 0 {
		session.TerminalRows = 24
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO tty_sessions 
			(id, asset_id, user_id, token, status, command, terminal_cols, terminal_rows, 
			 created_at, browser_ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	zap.L().Debug("Creating TTY session",
		zap.String("session_id", session.ID),
		zap.String("asset_id", session.AssetID),
		zap.String("user_id", session.UserID),
		zap.Int("cols", session.TerminalCols),
		zap.Int("rows", session.TerminalRows),
		zap.String("query", query))

	_, err := db.Exec(query,
		session.ID,
		session.AssetID,
		session.UserID,
		session.Token,
		session.Status,
		session.Command,
		session.TerminalCols,
		session.TerminalRows,
		session.CreatedAt,
		session.BrowserIP,
	)

	if err != nil {
		zap.L().Error("Failed to create TTY session",
			zap.String("session_id", session.ID),
			zap.String("asset_id", session.AssetID),
			zap.String("user_id", session.UserID),
			zap.String("query", query),
			zap.Error(err))
		return err
	}

	zap.L().Info("TTY session created",
		zap.String("session_id", session.ID),
		zap.String("asset_id", session.AssetID),
		zap.String("token_prefix", session.Token[:min(16, len(session.Token))]+"..."))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetTTYSessionByToken 通过Token获取会话
func GetTTYSessionByToken(token string) (*model.TTYSession, error) {
	var session model.TTYSession
	query := `
		SELECT id, asset_id, user_id, token, status, command, terminal_cols, terminal_rows,
		       created_at, connected_at, closed_at, browser_ip, agent_ip,
		       record_file, error_message
		FROM tty_sessions
		WHERE token = ? AND status != 'closed'
	`

	err := db.Get(&session, query, token)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetAgentTTYSessions 获取Agent需要处理的TTY会话
func GetAgentTTYSessions(assetID string) ([]model.TTYSession, error) {
	var sessions []model.TTYSession

	query := `
		SELECT id, asset_id, user_id, token, command, terminal_cols, terminal_rows, created_at
		FROM tty_sessions
		WHERE asset_id = ? AND status = 'pending'
		  AND created_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)
		ORDER BY created_at ASC
		LIMIT 10
	`

	err := db.Select(&sessions, query, assetID)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// CreateTTYToken 插入一次性 token（用于授权阶段）
func CreateTTYToken(token, userID, assetID string, cols, rows int, expireAt time.Time) error {
	query := `
		INSERT INTO tty_tokens 
			(token, user_id, asset_id, terminal_cols, terminal_rows, expire_at, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), 'pending')
	`
	_, err := db.Exec(query, token, userID, assetID, cols, rows, expireAt)
	if err != nil {
		zap.L().Error("CreateTTYToken failed", zap.String("token", token[:8]+"..."), zap.Error(err))
	}
	return err
}

// ConsumeTTYToken 原子消费 token（FOR UPDATE + 更新状态），防止重放
func ConsumeTTYToken(token string) (*model.TTYToken, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var t model.TTYToken
	err = tx.Get(&t, `
		SELECT id, token, user_id, asset_id, terminal_cols, terminal_rows, expire_at 
		FROM tty_tokens 
		WHERE token = ? AND status = 'pending' 
		FOR UPDATE`, token)
	if err != nil {
		return nil, err // 不存在或已被使用
	}

	if time.Now().After(t.ExpireAt) {
		return nil, errors.New("token expired")
	}

	_, err = tx.Exec(`UPDATE tty_tokens SET status = 'used', used_at = NOW() WHERE id = ?`, t.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}
