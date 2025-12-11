package task

import (
	"github.com/chiwen/server/internal/data/mysql"
	"go.uber.org/zap"
)

// TTYSessionCleanup 清理过期的TTY会话
func TTYSessionCleanup() {
	zap.L().Info("Starting TTY session cleanup")

	// 清理过期的pending会话
	if err := mysql.CleanupExpiredSessions(); err != nil {
		zap.L().Error("Failed to cleanup expired TTY sessions", zap.Error(err))
	}

	// 清理30分钟前已关闭的会话（可选）
	query := `
		DELETE FROM tty_sessions 
		WHERE status = 'closed' 
		  AND closed_at < DATE_SUB(NOW(), INTERVAL 30 MINUTE)
		LIMIT 1000
	`

	result, err := mysql.DB().Exec(query)
	if err != nil {
		zap.L().Error("Failed to cleanup closed TTY sessions", zap.Error(err))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		zap.L().Info("Cleaned up closed TTY sessions",
			zap.Int64("count", rowsAffected))
	}
}
