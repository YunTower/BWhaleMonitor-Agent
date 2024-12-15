package main

import (
	"agent/pkg/logger"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"os"
	"strings"
	"time"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
type MemoryInfo struct {
	Total   int `json:"total"`
	Used    int `json:"used"`
	Free    int `json:"free"`
	Percent int `json:"used_percent"`
}

var log *logger.Logger
var websocketApi = "ws://127.0.0.1:8097"
var key string = ""

const agentVersion = "0.0.1"

func init() {
	var err error
	log, err = logger.NewLogger("./logs")
	if err != nil {
		panic(err)
	}
}

func sendMessage(content interface{}, conn *websocket.Conn) {
	data, err := json.Marshal(content)
	if err != nil {
		log.Error("将内容转换为 JSON 时出错: %v", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Error("发送消息时出错: %v", err)
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

func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func main() {
	fmt.Println("================= 蓝鲸服务器探针 Agent v", agentVersion, " =================")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("主控WebSocket API: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return
	}
	//websocketApi = strings.TrimSpace(input)

	fmt.Print("通信密钥: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return
	}
	key = strings.TrimSpace(input)

	log.Info("API: %v KEY: %v", websocketApi, key)
	log.Info("正在尝试连接...")
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8097", nil)
	if err != nil {
		log.Error("连接失败: %v\n响应: %v", err)
		return
	}
	defer conn.Close()

	log.Success("WebSocket 连接成功")

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
		var jsonData map[string]interface{}
		err = json.Unmarshal(message, &jsonData)
		if err != nil {
			log.Error("解析JSON数据时出错:", err)
			continue
		}
		typeValue := jsonData["type"]
		switch typeValue {
		case "hello":
			content := Message{
				Type: "hi",
			}
			sendMessage(content, conn)
		case "auth":
			authData := map[string]string{
				"key": key,
			}
			content := Message{
				Type: "auth",
				Data: authData,
			}
			sendMessage(content, conn)
		}
	}
}
