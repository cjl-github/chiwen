package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chiwen/server/internal/data/model"
	"github.com/chiwen/server/internal/data/mysql"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// TTYService TTY相关服务
type TTYService struct{}

// NewTTYService 创建TTY服务实例
func NewTTYService() *TTYService {
	return &TTYService{}
}

// generateToken 生成一次性Token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// checkRateLimit 检查频率限制
func checkRateLimit(userID, assetID string) error {
	// 检查用户并发会话数
	userSessions, err := mysql.GetUserActiveSessionsCount(userID)
	if err != nil {
		return fmt.Errorf("failed to check user sessions: %w", err)
	}

	// 限制每个用户最多3个并发会话
	if userSessions >= 3 {
		return errors.New("user has too many active sessions (max: 3)")
	}

	// 检查机器并发会话数
	assetSessions, err := mysql.GetActiveSessionsCount(assetID)
	if err != nil {
		return fmt.Errorf("failed to check asset sessions: %w", err)
	}

	// 限制每台机器最多5个并发会话
	if assetSessions >= 5 {
		return errors.New("machine has too many active sessions (max: 5)")
	}

	return nil
}

// ValidateTTYToken 验证Token有效性
func (s *TTYService) ValidateTTYToken(token string) (*model.TTYSession, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	// 获取会话
	session, err := mysql.GetTTYSessionByToken(token)
	if err != nil {
		zap.L().Warn("Invalid TTY token", zap.String("token", token[:16]+"..."))
		return nil, errors.New("invalid or expired token")
	}

	// 检查Token是否过期（5分钟有效期）
	if time.Since(session.CreatedAt) > 5*time.Minute {
		// 标记为过期
		_ = mysql.UpdateTTYSessionStatus(session.ID, "closed")
		zap.L().Warn("TTY token expired",
			zap.String("session_id", session.ID),
			zap.String("token", token[:16]+"..."))
		return nil, errors.New("token expired")
	}

	// 检查会话状态
	if session.Status != "pending" && session.Status != "connected" {
		zap.L().Warn("TTY session not available",
			zap.String("session_id", session.ID),
			zap.String("status", session.Status))
		return nil, fmt.Errorf("session is not available (status: %s)", session.Status)
	}

	zap.L().Debug("TTY token validated",
		zap.String("session_id", session.ID),
		zap.String("asset_id", session.AssetID))

	return session, nil
}

// GetAgentTTYSessions 获取Agent需要处理的TTY会话
func (s *TTYService) GetAgentTTYSessions(assetID string) ([]model.TTYSession, error) {
	var sessions []model.TTYSession

	query := `
		SELECT id, asset_id, user_id, token, command, terminal_cols, terminal_rows, created_at
		FROM tty_sessions
		WHERE asset_id = ? AND status = 'pending'
		  AND created_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)
		ORDER BY created_at ASC
		LIMIT 10
	`

	err := mysql.DB().Select(&sessions, query, assetID)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("Get agent TTY sessions",
		zap.String("asset_id", assetID),
		zap.Int("count", len(sessions)))

	return sessions, nil
}

func getServerHost() string {
	// 优先从配置读取外部访问地址（推荐做法）
	if url := viper.GetString("server.external_url"); url != "" {
		// 去掉可能的前缀 wss:// 或 https://
		url = strings.TrimPrefix(url, "wss://")
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "ws://")
		url = strings.TrimPrefix(url, "http://")
		return strings.TrimRight(url, "/")
	}

	// 其次从配置读取监听端口，拼接 host
	host := viper.GetString("server.host")
	if host == "" {
		host = "localhost"
	}
	port := viper.GetInt("app.listen")
	if port == 0 {
		port = 8090
	}
	if port == 80 || port == 443 {
		return host
	}
	return fmt.Sprintf("localhost:%d", viper.GetInt("app.listen"))
}

// AuthorizeTTY 授权TTY访问
func (s *TTYService) AuthorizeTTY(assetID, userID, browserIP string, cols, rows int) (*model.TTYTokenInfo, error) {
	// 1. 参数校验
	if assetID == "" || userID == "" {
		return nil, errors.New("asset_id and user_id are required")
	}
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
	}

	// 2. 权限 + 状态检查
	// allowed, err := checkUserPermission(assetID, userID)
	// if err != nil {
	// 	return nil, err
	// }
	// if !allowed {
	// 	return nil, errors.New("user is not allowed to access this machine")
	// }
	// 当前所有审批通过的机器都可以直接连（用于快速测试）

	// 3. 并发会话限制
	if err := checkRateLimit(userID, assetID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// 4. 生成一次性 token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 5. 写入 tty_tokens 表（5分钟有效）
	if err := mysql.CreateTTYToken(token, userID, assetID, cols, rows, time.Now().Add(5*time.Minute)); err != nil {
		return nil, err
	}

	// 6. 构造 WebSocket URL（强烈建议用 external_url）
	wsHost := getServerHost() // 你后面我会给你这个函数
	wsURL := fmt.Sprintf("wss://%s/api/v1/tty/ws?token=%s", wsHost, token)

	zap.L().Info("TTY authorized success",
		zap.String("asset_id", assetID),
		zap.String("user_id", userID),
		zap.String("token_prefix", token[:12]+"..."))

	return &model.TTYTokenInfo{
		Token:   token,
		WsURL:   wsURL,
		Expires: 300,
	}, nil
}

// service/tty_service.go 中修改 checkUserPermission 函数
func checkUserPermission(assetID, userID string) (bool, error) {
	// 获取资产
	asset, err := mysql.GetAssetByID(assetID)
	if err != nil {
		return false, fmt.Errorf("asset not found: %w", err)
	}
	if asset.Status != "online" {
		return false, errors.New("machine is not online")
	}
	if asset.IsDeleted {
		return false, errors.New("machine has been deleted")
	}

	// 获取允许用户列表
	allowedUsers, err := asset.GetAllowedUsersArray()
	if err != nil {
		return false, fmt.Errorf("invalid allowed_users format: %w", err)
	}

	// 关键修改：支持通配符 `*`
	for _, allowed := range allowedUsers {
		if allowed == "*" || allowed == userID {
			zap.L().Debug("权限校验通过",
				zap.String("user_id", userID),
				zap.String("allowed", allowed),
				zap.String("asset_id", assetID))
			return true, nil
		}
	}

	// 无匹配
	zap.L().Warn("权限校验失败",
		zap.String("user_id", userID),
		zap.Int("allowed_count", len(allowedUsers)),
		zap.String("asset_id", assetID))
	return false, errors.New("user not in allowed users list")
}
