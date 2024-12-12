package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
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
	//_system := system.System{}
	//memoryInfo := MemoryInfo{
	//	Total:   _system.GetMemoryTotal(),
	//	Used:    _system.GetMemoryUsed(),
	//	Free:    _system.GetMemoryFree(),
	//	Percent: _system.GetMemoryUsedPercent(),
	//}
	//fmt.Println(memoryInfo)
	//fmt.Println(_system.GetIpv4())

	// 获取用户输入的内容
	// 处理命令行参数
	if len(os.Args) > 1 {
		fmt.Println("命令行参数:", os.Args[1])
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("请输入内容: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入时出错:", err)
			return
		}
		fmt.Println("您输入的内容:", input)
	}

}
