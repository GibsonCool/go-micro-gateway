package main

import (
	"fmt"
	"net"
)

func main() {
	// 1.监听端口
	listen, err := net.Listen("tcp", "0.0.0.0:9090")
	if err != nil {
		fmt.Println("获取监听失败,err:", err.Error())
		return
	}

	// 2.接受请求
	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("建立连接失败,err:", err.Error())
			return
		}

		// 3.创建协程
		go process(accept)
	}

}

func process(accept net.Conn) {
	defer accept.Close()
	for {
		buf := make([]byte, 128)
		n, err := accept.Read(buf)
		if err != nil {
			fmt.Println("接受消息失败,err:", err.Error())
			break
		}
		fmt.Println("接收到的消息：", string(buf[:n]))
	}
}
