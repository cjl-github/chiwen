package routes

import (
	"github.com/chiwen/server/internal/api/handler"
	"github.com/chiwen/server/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Setup() *gin.Engine {
	// 设置 Gin 模式
	mode := viper.GetString("app.mode")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// 创建处理器实例
	ttyHandler := handler.NewTTYHandler()

	api := r.Group("/api/v1")
	{
		// 原有的路由
		api.POST("/register", handler.RegisterHandler)
		api.POST("/approve", handler.ApproveHandler)
		api.POST("/heartbeat", handler.HeartbeatHandler)
		api.GET("/register/status", handler.RegisterStatusHandler)

		// 诊断接口
		api.GET("/diagnostic", handler.DiagnosticHandler)
		api.GET("/db-stats", handler.DatabaseStatsHandler)

		// TTY相关路由
		ttyGroup := api.Group("/tty")
		{
			ttyGroup.GET("/validate", ttyHandler.ValidateToken)
			// WebSocket路由
			ttyGroup.GET("/ws", handler.HandleWebSocket) // ← 改成这样！
		}

		// 资产相关路由
		assetsGroup := api.Group("/assets")
		{
			assetsGroup.GET("/:id/tty/authorize", ttyHandler.AuthorizeTTY)
		}

		// Agent相关路由
		agentGroup := api.Group("/agent")
		{
			// 旧的轮询接口（兼容旧版Agent可保留）
			agentGroup.GET("/tty/sessions", ttyHandler.GetAgentSessions)
			// 新增：Agent 长连接 WebSocket 入口
			agentGroup.GET("/tty/agent/ws", handler.AgentWebSocketHandler)
		}
	}
	return r
}
