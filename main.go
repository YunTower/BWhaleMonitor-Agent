package main

import (
	"agent/pkg/system"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type Content []byte
type MemoryInfo struct {
	Total   int `json:"total"`
	Used    int `json:"used"`
	Free    int `json:"free"`
	Percent int `json:"used_percent"`
}

func sendMessage(content Content, conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.TextMessage, content)
	if err != nil {
		log.Println("Error writing message:", err)
		return
	}
}

func main() {
	_system := system.System{}
	memoryInfo := MemoryInfo{
		Total:   _system.GetMemoryTotal(),
		Used:    _system.GetMemoryUsed(),
		Free:    _system.GetMemoryFree(),
		Percent: _system.GetMemoryUsedPercent(),
	}
	fmt.Println(memoryInfo)
	fmt.Println(_system.GetIpv4())
}
