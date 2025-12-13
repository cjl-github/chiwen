// internal/task/init.go
package task

import (
	"time"

	"go.uber.org/zap"
)

// StartBackgroundTasks 启动所有后台定时任务
func StartBackgroundTasks() {
	// 第一次立刻执行一次
	OfflineDetector()

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			OfflineDetector()
		}
	}()

	// TTY会话清理任务（每5分钟）
	ticker2 := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker2.Stop()
		for range ticker2.C {
			TTYSessionCleanup()
		}
	}()

	zap.L().Info("background tasks started (offline detector every 30 seconds)")
}
