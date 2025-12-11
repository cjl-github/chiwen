package task

import (
	"time"

	"github.com/chiwen/server/internal/data/mysql"
	"go.uber.org/zap"
)

const OfflineThreshold = 1 * time.Minute // 改为1分钟

func OfflineDetector() {
	cutoff := time.Now().Add(-OfflineThreshold)

	query := `
        UPDATE assets 
        SET status = 'offline' 
        WHERE status = 'online' 
          AND updated_at < ?
          AND is_deleted = 0
    `

	result, err := mysql.DB().Exec(query, cutoff)
	if err != nil {
		zap.L().Error("Offline detector failed",
			zap.String("query", query),
			zap.Time("cutoff", cutoff),
			zap.Error(err))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		zap.L().Info("Offline detector marked machines as offline",
			zap.Int64("count", rowsAffected),
			zap.Time("cutoff_time", cutoff))

		// 记录哪些机器被标记为离线
		var offlineIDs []string
		err = mysql.DB().Select(&offlineIDs,
			"SELECT id FROM assets WHERE status = 'offline' AND updated_at < ? AND is_deleted = 0",
			cutoff)

		if err != nil {
			zap.L().Warn("Failed to get offline machine IDs", zap.Error(err))
		} else if len(offlineIDs) > 0 {
			zap.L().Info("Machines marked offline",
				zap.Int("count", len(offlineIDs)),
				zap.Strings("ids", offlineIDs))
		}
	} else {
		zap.L().Debug("No machines need to be marked offline",
			zap.Time("cutoff_time", cutoff),
			zap.String("threshold", OfflineThreshold.String()))
	}
}
