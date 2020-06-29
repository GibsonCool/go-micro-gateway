package main

import (
	"fmt"
	"net"
)

func main() {
	// 1、监听服务
	listenUDP, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9090,
	})
	if err != nil {
		fmt.Println("监听失败，err:", err.Error())
		return
	}

	// 2.循环读取消息
	for {
		data := [1024]byte{}
		n, addr, err := listenUDP.ReadFromUDP(data[:])
		if err != nil {
			fmt.Println("接收数据失败，err:", err.Error())
			break
		}

		go func() {
			fmt.Printf("addr: %v , data: %v , count %v\n", addr, string(data[:n]), n)
			// 3.回复消息
			_, err := listenUDP.WriteToUDP([]byte("接收成功"), addr)
			if err != nil {
				fmt.Println("回写数据失败，err:", err.Error())
			}
		}()
	}

}
