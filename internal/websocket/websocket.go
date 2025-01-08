package websocket

import (
	"agent/internal/logger"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func Connect(api string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(api, nil)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	return conn, nil
}

func StartHeartbeat(conn *websocket.Conn, logger *logger.Logger) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		heartbeatMessage := Message{
			Type: "hello",
		}
		SendMessage(heartbeatMessage, conn, logger)
	}
}

func SendMessage(content interface{}, conn *websocket.Conn, logger *logger.Logger) {
	data, err := json.Marshal(content)
	if err != nil {
		logger.Error("将内容转换为 JSON 时出错: %v", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		logger.Error("发送消息时出错: %v", err)
		return
	}
}

func HandleDisconnect(conn *websocket.Conn, logger *logger.Logger) {
	if conn == nil {
		logger.Error("WebSocket 连接未建立")
		return
	}
	logger.Error("WebSocket 连接断开")
}
