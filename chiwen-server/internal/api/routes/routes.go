// routes/setup.go  （或者你原来叫 routes/router.go / routes/route.go，随你项目里叫什么）

package routes

import (
	"github.com/chiwen/server/internal/api/handler"
	"github.com/chiwen/server/internal/pkg/middleware"
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

	// ==================== 公开路由（无需登录） ====================
	api := r.Group("/api/v1")
	{
		// 登录接口（放在最上面，方便调试）
		api.POST("/login", handler.Login) // ← 你的 handler.Login 已经在 handler/auth.go 里实现了

		// 原有的公开接口
		api.POST("/register", handler.RegisterHandler)
		api.POST("/approve", handler.ApproveHandler)
		api.POST("/heartbeat", handler.HeartbeatHandler)
		api.GET("/register/status", handler.RegisterStatusHandler)

		// 诊断接口
		api.GET("/diagnostic", handler.DiagnosticHandler)
		api.GET("/db-stats", handler.DatabaseStatsHandler)

		// TTY 相关（部分需要登录，后面会移到 authGroup）
		ttyGroup := api.Group("/tty")
		{
			ttyGroup.GET("/validate", ttyHandler.ValidateToken)
			// WebSocket 路由（用户端
			ttyGroup.GET("/ws", handler.HandleWebSocket)
		}
	}

	// ==================== 需要登录的路由 ====================
	// 所有需要鉴权的接口统一放在这里
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.AuthRequired()) // ← 你的 JWT 中间件
	{
		// 资产相关（需要登录后才能看）
		assetsGroup := authGroup.Group("/assets")
		{
			// 你以后可以再加 ListAssets、DeleteAsset 等
			// assetsGroup.GET("", handler.ListAssets) // ← 建议你新增这个接口返回用户有权限的机器列表
			assetsGroup.GET("/:id/tty/authorize", ttyHandler.AuthorizeTTY)
		}

		// 如果以后还有用户管理、角色管理、审计日志等，都放这里
		// authGroup.GET("/users", handler.ListUsers)
		// authGroup.GET("/audit/logs", handler.GetAuditLogs)
	}

	// ==================== Agent 长连接（不需要用户登录） ====================
	agentGroup := r.Group("/api/v1/agent")
	{
		// 兼容旧版轮询
		agentGroup.GET("/tty/sessions", ttyHandler.GetAgentSessions)
		// 新版 Agent 长连接 WebSocket
		agentGroup.GET("/tty/agent/ws", handler.AgentWebSocketHandler)
	}

	return r
}
