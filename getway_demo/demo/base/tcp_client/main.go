package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	doSend()

}

func doSend() {
	// 1.连接服务器
	conn, err := net.Dial("tcp", "localhost:9090")
	defer conn.Close()
	if err != nil {
		fmt.Println("连接失败,err:", err.Error())
		return
	}

	// 2.读取命令行输入
	inputReader := bufio.NewReader(os.Stdin)
	for {
		// 3.不断读取消息，直到换行 \n
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("读取控制台消息失败,err:", err.Error())
			break
		}

		// 4.如果读取到 Q 则退出程序
		trimSpace := strings.TrimSpace(input)
		if trimSpace == "Q" {
			break
		}

		// 5.将命令号读取的到的消息发送给服务器
		_, err = conn.Write([]byte(trimSpace))
		if err != nil {
			fmt.Println("发送消息失败,err:", err.Error())
			break
		}
	}
}
