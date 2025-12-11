// internal/api/handler/websocket_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chiwen/server/internal/api/agent"
	"github.com/chiwen/server/internal/data/model"
	"github.com/chiwen/server/internal/data/mysql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var browserUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleWebSocket 处理浏览器端的 WebSocket 连接
func HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	// 消费一次性 token
	ttyToken, err := mysql.ConsumeTTYToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	// 创建会话记录
	session := &model.TTYSession{
		ID:           uuid.New().String(),
		AssetID:      ttyToken.AssetID,
		UserID:       ttyToken.UserID,
		Token:        token,
		Status:       "connected",
		Command:      "/bin/bash",
		TerminalCols: ttyToken.TerminalCols,
		TerminalRows: ttyToken.TerminalRows,
		BrowserIP:    c.ClientIP(),
		CreatedAt:    time.Now(),
		ConnectedAt:  time.Now(),
	}

	if err := mysql.CreateTTYSession(session); err != nil {
		zap.L().Error("创建会话失败", zap.Error(err))
		c.JSON(500, gin.H{"error": "create session failed"})
		return
	}

	// 升级为 WebSocket
	conn, err := browserUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("WebSocket 升级失败", zap.Error(err))
		return
	}
	defer conn.Close()

	// 查找 Agent 是否在线
	agentConn, ok := agent.AgentConns[session.AssetID]
	if !ok {
		conn.WriteJSON(gin.H{"type": "error", "message": "Agent 当前不在线，请稍后再试"})
		zap.L().Warn("Agent 不在线", zap.String("asset_id", session.AssetID))
		return
	}

	zap.L().Info("三方终端转发开始",
		zap.String("session_id", session.ID),
		zap.String("asset_id", session.AssetID),
		zap.String("user_id", session.UserID))

	// 通知 Agent 创建 PTY
	agentConn.WriteJSON(map[string]interface{}{
		"type":          "new_session",
		"id":            session.ID,
		"command":       "/bin/bash",
		"terminal_cols": session.TerminalCols,
		"terminal_rows": session.TerminalRows,
	})

	// 浏览器 → Server → Agent（输入 + resize）
	go func() {
		defer func() {
			agentConn.WriteJSON(map[string]interface{}{
				"type": "close_session", "session_id": session.ID,
			})
		}()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 尝试解析是否为控制消息
			var ctrl struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &ctrl) == nil && ctrl.Type == "resize" {
				var r struct {
					Cols int `json:"cols"`
					Rows int `json:"rows"`
				}
				if json.Unmarshal(data, &r) == nil {
					agentConn.WriteJSON(map[string]interface{}{
						"type":       "resize",
						"session_id": session.ID,
						"cols":       r.Cols,
						"rows":       r.Rows,
					})
				}
				continue
			}

			// 普通输入
			agentConn.WriteJSON(map[string]interface{}{
				"type":       "input",
				"session_id": session.ID,
				"data":       string(data),
			})
		}
	}()

	// Agent → Server → Browser（输出）
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var payload struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
			Data      string `json:"data"`
		}
		if json.Unmarshal(msg, &payload) == nil &&
			payload.Type == "output" &&
			payload.SessionID == session.ID {
			conn.WriteMessage(websocket.TextMessage, []byte(payload.Data))
		}
	}

	// 会话结束
	mysql.UpdateTTYSessionStatus(session.ID, "closed")
	zap.L().Info("会话已结束", zap.String("session_id", session.ID))
}
