package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chiwen/client/internal/api/handler"
	"github.com/chiwen/client/internal/api/routes"

	"github.com/chiwen/client/internal/service"
	"github.com/chiwen/client/pkg/config"
	"github.com/chiwen/client/pkg/logger"
	"go.uber.org/zap"
)

// NewServerCommand 返回一个 cobra.Command，表示启动 server 的命令。
// 这样可以在 main 中只负责解析/执行命令，保持 main.go 非常简洁。
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	return cmd
}

// run 包含启动服务的全部流程，并返回错误（供 cobra 处理）
func run() error {
	// -----------------------
	// 1. 初始化配置
	// -----------------------
	if err := config.Init(); err != nil {
		return fmt.Errorf("init config failed: %w", err)
	}

	// -----------------------
	// 2. 初始化日志
	// -----------------------
	if err := logger.InitLogger(); err != nil {
		return fmt.Errorf("init logger failed: %w", err)
	}
	zap.L().Info("Logger initialized")

	// -----------------------
	// 3. 客户端：确保凭证（注册 or 跳过）
	// -----------------------
	id, agentSecret, err := handler.Register()
	if err != nil {
		zap.L().Error("register failed", zap.Error(err))
		return err
	}

	zap.L().Info("Client registration OK or skipped",
		zap.String("client_id", id),
	)

	// -----------------------
	// 4. 启动心跳循环
	// -----------------------
	interval := viper.GetInt("client.heartbeat_interval")
	dataDir := viper.GetString("client.data_dir")

	zap.L().Info("Starting heartbeat loop",
		zap.String("id", id),
		zap.Int("interval", interval),
	)

	go service.HeartbeatLoop(
		id,
		agentSecret,
		interval,
		dataDir,
	)
	go service.StartAgentWebSocketLoop(id, agentSecret)

	// -----------------------
	// 5. 初始化路由
	// -----------------------
	router := routes.Setup() // 构建 gin 路由并返回 *gin.Engine（包含中间件和路由注册）

	// -----------------------
	// 6. HTTP Server
	// -----------------------
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.listen")), // 从配置读取端口，格式化为 ":8080" 这种形式
		Handler: router,                                         // 将 gin 引擎作为 Handler
		// 建议：可以在这里设置 ReadTimeout / WriteTimeout / IdleTimeout 等
	}

	// -----------------------
	// 7. 异步启动 HTTP server
	// -----------------------
	go func() {
		zap.L().Info("Server started", zap.String("addr", srv.Addr)) // 打印启动地址
		// ListenAndServe 会阻塞直到服务器关闭或发生错误
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 如果错误不是因为正常 Shutdown 导致的 ErrServerClosed，则认为是异常退出，记录致命日志
			zap.L().Fatal("listen error", zap.Error(err))
		}
	}()

	// -----------------------
	// 8. 优雅关闭
	// -----------------------
	quit := make(chan os.Signal, 1)                      // 创建接收信号的通道（缓冲 1 个）
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 监听 SIGINT (Ctrl+C) 和 SIGTERM（kill 默认）
	<-quit                                               // 阻塞直到收到上述信号之一

	zap.L().Info("Shutting down server...") // 收到退出信号，开始关闭流程

	// 创建一个超时 Context，用于控制 Shutdown 的最长等待时间（例如 5 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 确保 context 的 cancel 在函数退出时被调用，释放资源

	// 通过 http.Server.Shutdown 发起优雅关闭：等待正在处理的请求完成、停止接收新连接
	if err := srv.Shutdown(ctx); err != nil {
		// 如果在超时内关闭失败，记录错误（这里选择 Error 而非 Fatal，程序随后退出）
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting") // 最终退出日志
	return nil                     // 正常返回 nil（cobra 会认为命令成功）
}
