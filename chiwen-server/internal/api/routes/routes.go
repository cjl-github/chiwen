// routes/setup.go
package routes

import (
	"time"

	"github.com/gin-contrib/cors" // ← 务必保留这一行
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/chiwen/server/internal/api/handler"
	"github.com/chiwen/server/internal/pkg/middleware"
	"github.com/chiwen/server/pkg/logger"
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

	// ==================== 正确的 CORS 中间件（彻底解决 Failed to fetch）===================
	// 开发阶段直接放行所有来源，生产时改成 AllowOrigins 即可
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true, // ← 开发时用这个最快最稳
		// 上线后只需要把上面一行注释掉，打开下面这几行即可：
		// AllowOrigins: []string{
		// 	"http://localhost:5173",
		// 	"http://127.0.0.1:5173",
		// 	"https://your-domain.com",
		// },

		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// ================================================================================

	// 创建处理器实例
	ttyHandler := handler.NewTTYHandler()

	// ==================== 公开路由（无需登录） ====================
	api := r.Group("/api/v1")
	{
		// 登录接口（放在最上面，方便调试）
		api.POST("/login", handler.Login)

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
			// WebSocket 路由（用户端）
			ttyGroup.GET("/ws", handler.HandleWebSocket)
		}
	}

	// ==================== 需要登录的路由 ====================
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.AuthRequired()) // JWT 鉴权中间件
	{
		// 资产相关（需要登录后才能看）
		assetsGroup := authGroup.Group("/assets")
		{
			assetsGroup.GET("/:id/tty/authorize", ttyHandler.AuthorizeTTY)
		}
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
