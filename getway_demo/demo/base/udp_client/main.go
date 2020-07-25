package main

import (
	"fmt"
	"net"
)

func main() {
	// 1.连接服务器
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 9090,
	})
	if err != nil {
		fmt.Println("连接失败，err:", err.Error())
		return
	}

	for i := 0; i < 10; i++ {
		// 2.发送数据
		_, err := conn.Write([]byte("你好，服务器"))
		if err != nil {
			fmt.Println("发送失败，err:", err.Error())
			return
		}

		// 3.接受数据
		result := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(result)
		if err != nil {
			fmt.Println("接受服务器数据失败，err:", err.Error())
			return
		}
		fmt.Printf("接收到服务器的数据，addr: %v , data: %v\n", addr, string(result[:n]))
	}
}
