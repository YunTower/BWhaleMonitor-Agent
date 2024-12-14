package main

import (
	"agent/pkg/logger"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"time"
)

type Content []byte
type MemoryInfo struct {
	Total   int `json:"total"`
	Used    int `json:"used"`
	Free    int `json:"free"`
	Percent int `json:"used_percent"`
}

var log *logger.Logger

const websocketApi = "ws://0.0.0.0:8097"

func init() {
	var err error
	log, err = logger.NewLogger("./logs")
	if err != nil {
		panic(err)
	}
}

func sendMessage(content Content, conn *websocket.Conn) {
	if conn == nil {
		log.Error("WebSocket 连接未建立")
		return
	}
	err := conn.WriteMessage(websocket.TextMessage, content)
	if err != nil {
		log.Error(fmt.Sprintf("Error writing message: %v", err))
		return
	}
}

func handleDisconnect(conn *websocket.Conn) {
	if conn == nil {
		log.Error("WebSocket 连接未建立")
		return
	}
	log.Error("WebSocket 连接断开")
	reconnect(conn)
}

func reconnect(conn *websocket.Conn) {
	maxReconnectAttempts := 3
	attempts := 0

	for {
		if attempts >= maxReconnectAttempts {
			log.Error("已达到最大重连次数，请检测主控与被控直接的网络连接状态")
			return
		} else {
			attempts++
			log.Info("第%v次尝试重新连接...", attempts)
			time.Sleep(5 * time.Second)

			newConn, resp, err := websocket.DefaultDialer.Dial(websocketApi, nil)
			if err != nil {
				log.Error(fmt.Sprintf("重新连接失败: %v\n响应: %v", err, resp))
				continue
			}

			log.Success("重新连接成功")
			defer newConn.Close()

			content := Content(`{"type":"hello"}`)
			sendMessage(content, newConn)

			for {
				_, message, err := newConn.ReadMessage()
				if err != nil {
					log.Error(fmt.Sprintf("读取消息时出错: %v", err))
					handleDisconnect(newConn)
					break
				}
				log.Info(fmt.Sprintf("接收到消息: %s", message))
			}
		}
	}
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(websocketApi, nil)
	if err != nil {
		log.Error("连接失败: %v\n响应: %v", err)
		return
	}
	defer conn.Close()

	log.Success("WebSocket 连接成功")

	content := Content(`{"type":"hello"}`)
	sendMessage(content, conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				log.Warn("连接已关闭")
			} else {
				log.Error(fmt.Sprintf("读取消息时出错: %v", err))
			}
			handleDisconnect(conn)
			break
		}
		log.Info(fmt.Sprintf("接收到消息: %s", message))
	}
}
