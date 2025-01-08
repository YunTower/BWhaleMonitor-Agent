package main

import (
	"agent/config"
	"agent/internal/reporter"
	"agent/internal/websocket"
	"os"
)

const agentVersion = "0.0.1"

func main() {
	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化日志
	logger := config.InitLogger(cfg.LogPath)

	// 初始化系统信息
	system := config.InitSystem()

	// 初始化WebSocket连接
	conn, err := websocket.Connect(cfg.WebsocketAPI)
	if err != nil {
		logger.Error("连接失败: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	logger.Success("WebSocket 连接成功")

	// 定时心跳
	go websocket.StartHeartbeat(conn, logger)

	// 消息处理
	reporter.StartReporter(conn, system, logger, cfg)
}
