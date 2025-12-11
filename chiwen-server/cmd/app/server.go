package app

import (
	"context"   // 用于控制优雅关闭时的超时 Context
	"fmt"       // 格式化字符串
	"net/http"  // HTTP server 类型和常量
	"os"        // 操作系统相关（退出码等）
	"os/signal" // 捕获系统信号
	"syscall"   // 系统调用常量（SIGINT、SIGTERM）
	"time"      // 时间相关（超时、间隔）

	"github.com/spf13/cobra" // cobra：命令行框架，用来构建可执行命令
	"github.com/spf13/viper" // viper：配置管理库

	"github.com/chiwen/server/internal/api/routes" // 项目内部路由构建
	"github.com/chiwen/server/internal/data/mysql" // 项目内部 MySQL 初始化封装
	"github.com/chiwen/server/internal/task"
	"github.com/chiwen/server/pkg/config" // 项目配置初始化（包装 viper）
	"github.com/chiwen/server/pkg/logger" // 项目日志初始化（包装 zap）

	"go.uber.org/zap" // zap 日志库
)

// NewServerCommand 返回一个 cobra.Command，表示启动 server 的命令。
// 这样可以在 main 中只负责解析/执行命令，保持 main.go 非常简洁。
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",           // 命令名字：server
		Short: "Start the server", // 简短描述（在帮助中显示）
		RunE: func(cmd *cobra.Command, args []string) error { // RunE 支持返回 error（便于统一错误处理）
			return run() // 执行实际启动逻辑（抽离到 run()，便于测试）
		},
	}
	return cmd // 返回 cobra 命令实例
}

// run 包含启动服务的全部流程，并返回错误（供 cobra 处理）
func run() error {
	// 1. init config
	if err := config.Init(); err != nil { // 初始化配置（例如读取 configs/config.yaml）
		return fmt.Errorf("init config failed: %w", err) // 包装并向上返回错误
	}

	// 2. init logger
	if err := logger.InitLogger(); err != nil { // 初始化日志（zap/lumberjack 等）
		return fmt.Errorf("init logger failed: %w", err)
	}
	zap.L().Info("Logger initialized") // 全局 logger 就绪，打印一条信息

	// 3. init mysql
	if err := mysql.InitDB(); err != nil { // 初始化 MySQL 连接池（sqlx 或 database/sql）
		return fmt.Errorf("init mysql failed: %w", err)
	}
	defer mysql.Close() // 程序退出时关闭数据库连接（defer 在 run 返回时执行）

	// 新增：启动所有后台任务（离线检测、后续可以加更多定时任务）
	// 放在这里最合适：所有依赖（logger、mysql）都已初始化，HTTP 服务还没完全挡住主协程
	task.StartBackgroundTasks()

	// 4. router
	router := routes.Setup() // 构建 gin 路由并返回 *gin.Engine（包含中间件和路由注册）

	// 5. HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.listen")), // 从配置读取端口，格式化为 ":8080" 这种形式
		Handler: router,                                         // 将 gin 引擎作为 Handler
		// 建议：可以在这里设置 ReadTimeout / WriteTimeout / IdleTimeout 等
	}

	// 6. run server（在单独 goroutine 中运行 ListenAndServe）
	go func() {
		zap.L().Info("Server started", zap.String("addr", srv.Addr)) // 打印启动地址
		// ListenAndServe 会阻塞直到服务器关闭或发生错误
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 如果错误不是因为正常 Shutdown 导致的 ErrServerClosed，则认为是异常退出，记录致命日志
			zap.L().Fatal("listen error", zap.Error(err))
		}
	}()

	// 7. graceful shutdown（优雅关闭）
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
