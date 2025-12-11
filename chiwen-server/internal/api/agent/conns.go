// internal/agent/conns.go
package agent

import "github.com/gorilla/websocket"

// AgentConns 全局保存所有在线的 Agent 长连接
var AgentConns = make(map[string]*websocket.Conn)
