// internal/task/agent_offline_detector.go
package task

import (
	"time"

	"github.com/chiwen/server/internal/data/mysql"
	"go.uber.org/zap"
)

func init() {
	go func() {
		for range time.Tick(30 * time.Second) {
			if n, err := mysql.MarkAssetOfflineIfNoPing(); err == nil && n > 0 {
				zap.L().Info("Agent offline detected", zap.Int64("count", n))
			}
		}
	}()
}
