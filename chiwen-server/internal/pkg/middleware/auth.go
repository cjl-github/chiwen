// internal/middleware/middleware.go   （文件名随便你，反正就是放中间件的那一个）
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/chiwen/server/internal/pkg/utils"
)

// ==================== CORS 中间件（新增） ====================
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 开发阶段直接放行所有域名，最快最有效
		AllowAllOrigins: true,

		// 如果以后要收紧，改成下面这段即可
		// AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:3000"},

		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// ==================== 原来的鉴权中间件（完全不动） ====================
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "缺少 Authorization"})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "无效的 token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}
