package config

import (
	"agent/internal/logger"
	"agent/internal/system"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	WebsocketAPI string `json:"websocket"`
	Key          string `json:"key"`
	LogPath      string `json:"log_path"`
}

func LoadConfig() Config {
	var cfg Config
	// 检查是否存在agent.lock.json文件
	_, err := os.Stat("agent.lock.json")
	if err == nil {
		file, err := os.ReadFile("agent.lock.json")
		if err != nil {
			fmt.Println("读取锁定文件时出错:", err)
			os.Exit(1)
		}
		err = json.Unmarshal(file, &cfg)
		if err != nil {
			fmt.Println("解析JSON数据时出错:", err)
			os.Exit(1)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("主控WebSocket API: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入时出错:", err)
			os.Exit(1)
		}
		cfg.WebsocketAPI = strings.TrimSpace(input)

		fmt.Print("通信密钥: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入时出错:", err)
			os.Exit(1)
		}
		cfg.Key = strings.TrimSpace(input)

		// 创建锁定文件
		file, err := os.Create("agent.lock.json")
		if err != nil {
			fmt.Println("创建文件时出错:", err)
			os.Exit(1)
		}
		// 使用json格式写入密钥
		_, err = file.Write([]byte(`{"websocket":"` + cfg.WebsocketAPI + `","key":"` + cfg.Key + `"}`))
		if err != nil {
			fmt.Println("写入文件时出错:", err)
			os.Exit(1)
		}
		defer file.Close()
		fmt.Println("初始化成功，已生成锁定文件（如需重新初始化，请删除锁定文件）")
	}

	cfg.LogPath = "./logs"
	return cfg
}

func InitLogger(logPath string) *logger.Logger {
	logger, err := logger.NewLogger(logPath)
	if err != nil {
		fmt.Println("初始化日志时出错:", err)
		os.Exit(1)
	}
	return logger
}

func InitSystem() *system.System {
	return &system.System{}
}
