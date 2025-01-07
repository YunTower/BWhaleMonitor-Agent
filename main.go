package main

import (
	"agent/pkg/logger"
	"agent/pkg/system"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type CpuInfo struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	LogicCount int    `json:"logic_count"`
	Count      int    `json:"count"`
}

type MemoryInfo struct {
	Total   int `json:"total"`
	Used    int `json:"used"`
	Free    int `json:"free"`
	Percent int `json:"used_percent"`
}

type DiskInfo struct {
	Path        string  `json:"path"`
	Total       int     `json:"total"`
	Free        int     `json:"free"`
	Used        int     `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

var log *logger.Logger
var websocketApi = "ws://127.0.0.1:8097"
var key = ""
var isAgentInit = false
var ipv4 string

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

func marshalJSON(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("序列化数据时出错: %v", err)
	}
	return jsonData, err
}

func main() {
	/**
	 * 检查初始化状态
	 */
	// 检查是否存在agent.lock.json文件
	_, err := os.Stat("agent.lock.json")
	if err == nil {
		isAgentInit = true
	}

	/**
	 * 初始化Agent
	 */
	fmt.Printf("\033[32m================= 蓝鲸服务器探针 Agent v%v =================\033[0m\n", agentVersion)
	fmt.Printf("Github: https://github.com/YunTower/BWhaleMonitor\n")
	fmt.Printf("项目不断更新中，觉得好用可以点个Star~\n")
	fmt.Printf("被控端初始化成功后会自动以monitor用户运行\n")
	fmt.Printf("在服务器上可以通过\"bwmonitor\"命令快速调出控制面板\n")
	fmt.Printf("\033[32m======================== 初始化被控端 =========================\033[0m\n")
	if !isAgentInit {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("主控WebSocket API: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入时出错:", err)
			return
		}
		websocketApi = strings.TrimSpace(input)

		fmt.Print("通信密钥: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入时出错:", err)
			return
		}
		key = strings.TrimSpace(input)
	} else {
		log.Info("检测到已存在锁定文件，正在尝试读取...")
		file, err := os.ReadFile("agent.lock.json")
		if err != nil {
			log.Error("读取锁定文件时出错:", err)
			return
		}
		var jsonData map[string]interface{}
		err = json.Unmarshal(file, &jsonData)
		if err != nil {
			log.Error("解析JSON数据时出错:", err)
			return
		}
		websocketApi = jsonData["websocket"].(string)
		key = jsonData["key"].(string)
	}
	fmt.Printf("\033[32m====================== 被控端初始化成功 =======================\033[0m\n")
	log.Info("API: %v KEY: %v", websocketApi, key)
	system := system.System{}
	ipv4 := system.GetIpv4(log)
	log.Info("本机Ipv4: %v", ipv4)
	log.Info("正在尝试连接...")

	/**
	 * websocket
	 */
	conn, _, err := websocket.DefaultDialer.Dial(websocketApi, nil)
	if err != nil {
		log.Error("连接失败: %v\n响应: %v", err)
		return
	}
	defer conn.Close()

	log.Success("WebSocket 连接成功")

	/**
	 * 定时心跳
	 */
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			// 发送心跳消息
			heartbeatMessage := Message{
				Type: "hello",
			}
			sendMessage(heartbeatMessage, conn)
		}
	}()

	/*
	 * 消息处理
	 */
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

		log.Info(fmt.Sprintf("收到消息: %s", message))
		var jsonData map[string]interface{}
		err = json.Unmarshal(message, &jsonData)
		if err != nil {
			log.Error("解析JSON数据时出错:", err)
			continue
		}

		typeValue := jsonData["type"]
		statusValue, statusExists := jsonData["status"].(string)
		messageValue, messageExists := jsonData["message"].(string)

		if statusExists && typeValue == "auth" && statusValue == "success" {
			if !isAgentInit {
				// 创建锁定文件
				file, err := os.Create("agent.lock.json")
				if err != nil {
					fmt.Println("创建文件时出错:", err)
					return
				}
				// 使用json格式写入密钥
				_, err = file.Write([]byte(`{"websocket":"` + websocketApi + `","key":"` + key + `"}`))
				if err != nil {
					fmt.Println("写入文件时出错:", err)
					return
				}
				defer file.Close()
				log.Success("初始化成功，已生成锁定文件（如需重新初始化，请删除锁定文件）")
				isAgentInit = true
			}
		}

		if statusExists && messageExists {
			if statusValue != "success" {
				log.Warn("[%        v][%v] %v", typeValue, statusValue, messageValue)
			} else {
				log.Success("[%v][%v] %v", typeValue, statusValue, messageValue)
			}
		} else {
			if !statusExists {
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
				case "info":
					cpuInfo, _ := marshalJSON(system.GetCpuInfo())
					memoryInfo := map[string]int{
						"total": system.GetMemoryTotal(),
					}
					memoryTotal, _ := marshalJSON(memoryInfo)

					var diskData []DiskInfo
					for _, item := range system.GetDiskAllPart() {
						info := system.GetDiskUsage(item.Device)
						diskData = append(diskData, DiskInfo{
							Path:        item.Device,
							Total:       int(info.Total),
							Free:        int(info.Free),
							Used:        int(info.Used),
							UsedPercent: info.UsedPercent,
						})
					}

					DiskList, _ := marshalJSON(diskData)

					systemInfo := map[string]interface{}{
						"cpu":    json.RawMessage(cpuInfo),
						"memory": json.RawMessage(memoryTotal),
						"disk":   json.RawMessage(DiskList),
						"os":     runtime.GOOS,
						"arch":   runtime.GOARCH,
					}

					content := Message{
						Type: "info",
						Data: systemInfo,
					}
					sendMessage(content, conn)

				default:
					log.Warn("未知的消息类型:", typeValue)
				}
			}
		}
	}
}
