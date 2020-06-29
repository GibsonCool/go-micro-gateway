package main

import (
	"log"
	"net/http"
	"time"
)

var Addr = ":1210"

func main() {
	// 1.创建路由器
	mux := http.NewServeMux()

	// 2.设置路由规则
	mux.HandleFunc("/bye", sayBye)
	mux.HandleFunc("/hello", sayHello)

	// 3.创建服务器
	server := &http.Server{
		Addr:         Addr,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}
	log.Println("开启 http 服务，地址：" + Addr)
	log.Fatal(server.ListenAndServe())
}

func sayHello(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("hello world from http"))
}

func sayBye(writer http.ResponseWriter, request *http.Request) {
	time.Sleep(1 * time.Second)
	writer.Write([]byte("再见，这是 http 服务"))
}
