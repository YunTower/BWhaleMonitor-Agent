package reporter

import (
	"agent/config"
	"agent/internal/logger"
	"agent/internal/system"
	"agent/internal/websocket"
	"encoding/json"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"io"
	"runtime"
	"time"
)

type CpuInfo struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	LogicCount int    `json:"logic_count"`
	Count      int    `json:"count"`
}

type MemoryIo struct {
	Total       int `json:"total"`
	Used        int `json:"used"`
	Free        int `json:"free"`
	UsedPercent int `json:"used_percent"`
}

type DiskIo struct {
	Path        string  `json:"path"`
	Total       int     `json:"total"`
	Free        int     `json:"free"`
	Used        int     `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

func StartReporter(conn *websocket2.Conn, system *system.System, logger *logger.Logger, cfg config.Config) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				logger.Warn("连接已关闭")
			} else {
				logger.Error(fmt.Sprintf("读取消息时出错: %v，消息内容：%v", err, message))
			}
			websocket.HandleDisconnect(conn, logger)
			break
		}

		logger.Info(fmt.Sprintf("收到消息: %s", message))
		var jsonData map[string]interface{}
		err = json.Unmarshal(message, &jsonData)
		if err != nil {
			logger.Error("解析JSON数据时出错:", err)
			continue
		}

		typeValue := jsonData["type"]
		statusValue, statusExists := jsonData["status"].(string)
		messageValue, messageExists := jsonData["message"].(string)

		if statusExists && typeValue == "auth" && statusValue == "success" {
			logger.Success("认证成功")
		}

		if statusExists && messageExists {
			if statusValue != "success" {
				logger.Warn("[%v][%v] %v", typeValue, statusValue, messageValue)
			} else {
				logger.Success("[%v][%v] %v", typeValue, statusValue, messageValue)
			}

			if typeValue == "info" && statusValue == "success" {
				// 开启IO定时上报
				ticker := time.NewTicker(1 * time.Minute)
				defer ticker.Stop()
				go func() {
					for range ticker.C {
						reportSystemInfo(conn, system, logger)
					}
				}()
			}

		} else {
			if !statusExists {
				switch typeValue {
				case "hello":
					heartbeatMessage := websocket.Message{
						Type: "hi",
					}
					websocket.SendMessage(heartbeatMessage, conn, logger)
					break
				case "auth":
					authData := map[string]string{
						"type": "server",
						"key":  cfg.Key,
					}
					authMessage := websocket.Message{
						Type: "auth",
						Data: authData,
					}
					websocket.SendMessage(authMessage, conn, logger)
					break
				case "info":
					cpuInfo, _ := marshalJSON(system.GetCpuInfo(), logger)
					memoryInfo := map[string]int{
						"total": system.GetMemoryTotal(),
					}
					memoryTotal, _ := marshalJSON(memoryInfo, logger)

					var diskData []DiskIo
					for _, item := range system.GetDiskAllPart() {
						info := system.GetDiskUsage(item.Device)
						diskData = append(diskData, DiskIo{
							Path:        item.Device,
							Total:       int(info.Total),
							Free:        int(info.Free),
							Used:        int(info.Used),
							UsedPercent: info.UsedPercent,
						})
					}

					DiskList, _ := marshalJSON(diskData, logger)

					systemInfo := map[string]interface{}{
						"cpu":    json.RawMessage(cpuInfo),
						"memory": json.RawMessage(memoryTotal),
						"disk":   json.RawMessage(DiskList),
						"os":     runtime.GOOS,
						"arch":   runtime.GOARCH,
					}

					infoMessage := websocket.Message{
						Type: "info",
						Data: systemInfo,
					}
					websocket.SendMessage(infoMessage, conn, logger)
					break
				default:
					logger.Warn("未知的消息类型:", typeValue)
				}
			}
		}
	}
}

func reportSystemInfo(conn *websocket2.Conn, system *system.System, logger *logger.Logger) {
	cpuInfo, _ := marshalJSON(system.GetCpuUsedPercent(), logger)
	memoryInfo := map[string]int{
		"total": system.GetMemoryTotal(),
	}
	memoryTotal, _ := marshalJSON(memoryInfo, logger)

	var diskData []DiskIo
	for _, item := range system.GetDiskAllPart() {
		info := system.GetDiskUsage(item.Device)
		diskData = append(diskData, DiskIo{
			Path:        item.Device,
			Total:       int(info.Total),
			Free:        int(info.Free),
			Used:        int(info.Used),
			UsedPercent: info.UsedPercent,
		})
	}

	DiskList, _ := marshalJSON(diskData, logger)

	systemInfo := map[string]interface{}{
		"cpu":    json.RawMessage(cpuInfo),
		"memory": json.RawMessage(memoryTotal),
		"disk":   json.RawMessage(DiskList),
		"os":     runtime.GOOS,
		"arch":   runtime.GOARCH,
	}

	content := websocket.Message{
		Type: "info",
		Data: systemInfo,
	}
	websocket.SendMessage(content, conn, logger)
}

func marshalJSON(data interface{}, logger *logger.Logger) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {

		logger.Error("序列化数据时出错: %v", err)
	}
	return jsonData, err
}
