package routes

import (
	"net/http"

	"github.com/chiwen/client/pkg/logger"
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

	// 注册路由
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	return r
}
