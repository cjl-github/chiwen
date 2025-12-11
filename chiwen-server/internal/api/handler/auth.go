package handler

import (
	"net/http"
	"time"

	"github.com/chiwen/server/internal/data/model"
	"github.com/chiwen/server/internal/data/mysql"
	"github.com/chiwen/server/internal/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	var user model.User
	// 使用sqlx查询用户
	query := `SELECT id, username, password_hash, name, email, phone, is_active, is_admin, 
	                 ldap_dn, created_at, updated_at, last_login_at, last_login_ip 
	          FROM users 
	          WHERE username = ? AND is_active = 1`

	err := mysql.DB().Get(&user, query, req.Username)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Warn("登录失败：用户不存在或被禁用")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logrus.WithField("username", req.Username).Warn("登录失败：密码错误")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 token 失败"})
		return
	}

	// 更新登录信息
	updateQuery := `UPDATE users SET last_login_at = ?, last_login_ip = ? WHERE id = ?`
	_, err = mysql.DB().Exec(updateQuery, time.Now(), c.ClientIP(), user.ID)
	if err != nil {
		logrus.WithField("username", user.Username).Error("更新登录信息失败")
		// 不返回错误，因为登录已经成功
	}

	logrus.WithFields(logrus.Fields{
		"username": user.Username,
		"ip":       c.ClientIP(),
	}).Info("登录成功")

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}
