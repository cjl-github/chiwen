// internal/api/handler/agent_ws_handler.go
package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/chiwen/server/internal/api/agent"
	"github.com/chiwen/server/internal/data/mysql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var agentUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 改这里！
}

func AgentWebSocketHandler(c *gin.Context) {
	assetID := c.Query("asset_id")
	tsStr := c.Query("ts")
	sig := c.Query("sig")

	if assetID == "" || tsStr == "" || sig == "" {
		c.JSON(400, gin.H{"error": "missing auth params"})
		return
	}

	ts, _ := strconv.ParseInt(tsStr, 10, 64)
	if err := validateAgentAuth(assetID, ts, tsStr, sig); err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	conn, err := agentUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("Agent WS upgrade failed", zap.Error(err))
		return
	}

	// 注册 Agent 连接
	agent.AgentConns[assetID] = conn

	defer func() {
		delete(agent.AgentConns, assetID)
		mysql.UpdateAssetStatus(assetID, "offline")
		mysql.RemoveAgentConnection(assetID)
		conn.Close()
		zap.L().Info("Agent 已断开", zap.String("asset_id", assetID))
	}()

	// 记录连接时间
	mysql.RegisterAgentConnection(assetID, uuid.New().String(), c.ClientIP())

	// 欢迎消息
	conn.WriteJSON(gin.H{"type": "welcome", "message": "agent connected"})

	// 设置 pong 处理器
	conn.SetPongHandler(func(string) error {
		mysql.UpdateAgentPing(assetID)
		return nil
	})

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func validateAgentAuth(assetID string, ts int64, tsStr, signature string) error {
	if time.Now().Unix()-ts > 120 {
		return errors.New("timestamp expired")
	}

	secret, err := mysql.GetAgentSecretKeyByID(assetID)
	if err != nil || secret == "" {
		return errors.New("invalid asset or secret")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(assetID + tsStr))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return errors.New("invalid signature")
	}

	mysql.UpdateAssetStatus(assetID, "online")
	return nil
}
