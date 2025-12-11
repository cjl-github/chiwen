// internal/service/agent_ws.go
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// TTYSessionFromServer 服务端推送的会话结构
type TTYSessionFromServer struct {
	ID           string `json:"id"`
	AssetID      string `json:"asset_id"`
	UserID       string `json:"user_id"`
	Token        string `json:"token"`
	Command      string `json:"command"`
	TerminalCols int    `json:"terminal_cols"`
	TerminalRows int    `json:"terminal_rows"`
	BrowserIP    string `json:"browser_ip"`
}

// SessionPair 记录正在运行的会话
type SessionPair struct {
	Pty    *os.File
	Cmd    *exec.Cmd
	Cols   int
	Rows   int
	Closed bool
}

var (
	activeSessions = make(map[string]*SessionPair)
	sessionMu      sync.RWMutex
)

// StartAgentWebSocketLoop 启动 Agent 长连接（自动重连 + ping）
// 只需要调用一次！
func StartAgentWebSocketLoop(assetID, agentSecret string) {
	go func() {
		for {
			if err := connectAgentWS(assetID, agentSecret); err != nil {
				zap.L().Error("Agent WebSocket 断开，5秒后重连", zap.Error(err))
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func connectAgentWS(assetID, agentSecret string) error {
	// 构造带签名的 URL
	ts := fmt.Sprintf("%d", time.Now().Unix())
	msg := assetID + ts
	sig := hmacBase64Sign([]byte(agentSecret), []byte(msg))

	proto := "ws"
	if viper.GetString("server.protocol") == "https" {
		proto = "wss"
	}
	url := fmt.Sprintf("%s://%s:%d/api/v1/agent/tty/agent/ws?asset_id=%s&ts=%s&sig=%s",
		proto,
		viper.GetString("server.host"),
		viper.GetInt("server.port"),
		assetID, ts, sig)

	zap.L().Info("Agent 正在连接 WebSocket", zap.String("url", url))

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	zap.L().Info("Agent WebSocket 已连接", zap.String("asset_id", assetID))

	// 心跳
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}()

	// 主消息循环
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &base); err != nil {
			continue
		}

		switch base.Type {
		case "welcome":
			zap.L().Info("收到服务端欢迎消息")

		case "new_session":
			var session TTYSessionFromServer
			if err := json.Unmarshal(message, &session); err != nil {
				zap.L().Error("解析会话失败", zap.Error(err))
				continue
			}
			zap.L().Info("收到新会话请求",
				zap.String("session_id", session.ID),
				zap.Int("cols", session.TerminalCols),
				zap.Int("rows", session.TerminalRows))

			go handleTTYSession(conn, session)

		default:
			zap.L().Debug("未知消息类型", zap.String("type", base.Type))
		}
	}
}

// handleTTYSession 启动 pty 并双向转发
func handleTTYSession(agentConn *websocket.Conn, session TTYSessionFromServer) {
	cmd := exec.Command("/bin/bash")
	if session.Command != "" && session.Command != "/bin/bash" {
		cmd = exec.Command("sh", "-c", session.Command)
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		zap.L().Error("pty.Start 失败", zap.Error(err))
		return
	}
	defer ptmx.Close()

	// 设置初始大小
	_ = pty.Setsize(ptmx, &pty.Winsize{
		Cols: uint16(session.TerminalCols),
		Rows: uint16(session.TerminalRows),
	})

	// 注册会话
	sessionMu.Lock()
	activeSessions[session.ID] = &SessionPair{
		Pty:  ptmx,
		Cmd:  cmd,
		Cols: session.TerminalCols,
		Rows: session.TerminalRows,
	}
	sessionMu.Unlock()

	zap.L().Info("PTY 已启动", zap.String("session_id", session.ID))

	// Agent → Server → Browser（输出）
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				break
			}
			payload := map[string]interface{}{
				"type":       "output",
				"session_id": session.ID,
				"data":       string(buf[:n]),
			}
			if err := agentConn.WriteJSON(payload); err != nil {
				break
			}
		}
		closePTY(session.ID)
	}()

	// Browser → Server → Agent（输入 + resize）
	// Browser → Server → Agent（输入 + resize + close）
	for {
		_, data, err := agentConn.ReadMessage()
		if err != nil {
			break
		}

		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &base); err != nil {
			continue
		}

		switch base.Type {
		case "input":
			var input struct {
				SessionID string `json:"session_id"`
				Data      string `json:"data"`
			}
			if err := json.Unmarshal(data, &input); err == nil && input.SessionID == session.ID {
				writeToPTY(session.ID, []byte(input.Data))
			}

		case "resize":
			var r struct {
				SessionID string `json:"session_id"`
				Cols      int    `json:"cols"`
				Rows      int    `json:"rows"`
			}
			if err := json.Unmarshal(data, &r); err == nil && r.SessionID == session.ID {
				resizePTY(session.ID, r.Cols, r.Rows)
			}

		case "close_session":
			var c struct {
				SessionID string `json:"session_id"`
			}
			if err := json.Unmarshal(data, &c); err == nil && c.SessionID == session.ID {
				zap.L().Info("会话被主动关闭", zap.String("session_id", session.ID))
				return
			}
		}
	}

	// 命令结束或异常
	cmd.Wait()
	closePTY(session.ID)
}

// 写入输入
func writeToPTY(sessionID string, data []byte) {
	sessionMu.RLock()
	sp := activeSessions[sessionID]
	sessionMu.RUnlock()
	if sp != nil && !sp.Closed {
		sp.Pty.Write(data)
	}
}

// 调整终端大小
func resizePTY(sessionID string, cols, rows int) {
	sessionMu.RLock()
	sp := activeSessions[sessionID]
	sessionMu.RUnlock()
	if sp != nil {
		pty.Setsize(sp.Pty, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
		sp.Cols, sp.Rows = cols, rows
	}
}

// 关闭会话
func closePTY(sessionID string) {
	sessionMu.Lock()
	sp := activeSessions[sessionID]
	if sp != nil {
		sp.Closed = true
		if sp.Cmd.Process != nil {
			sp.Cmd.Process.Kill()
		}
		sp.Pty.Close()
		delete(activeSessions, sessionID)
	}
	sessionMu.Unlock()
}
