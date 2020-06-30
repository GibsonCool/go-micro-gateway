package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
	真正的运作服务器，监听 2003、2004 端口用作被代理服务器
*/
func main() {
	rs1 := &RealServer{"127.0.0.1:2003"}
	rs1.Run()
	rs2 := &RealServer{"127.0.0.1:2004"}
	rs2.Run()

	// 监听关闭信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

type RealServer struct {
	Addr string
}

func (r *RealServer) Run() {
	fmt.Println("开启 HTTP 服务，监听地址：", r.Addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.HelloHandler)
	server := &http.Server{
		Addr:         r.Addr,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()
}

func (r *RealServer) HelloHandler(writer http.ResponseWriter, request *http.Request) {
	// 返回访问地址路径
	uPath := fmt.Sprintf("http://%s%s\n", r.Addr, request.URL.Path)
	realIp := fmt.Sprintf("RemoteAddr=%s,X-Forwarded-For=%v,X-Real-Ip=%v\n",
		request.RemoteAddr,
		request.Header.Get("X-Forwarded-For"),
		request.Header.Get("X-Real-Ip"))
	header := fmt.Sprintf("headers=%v\n", request.Header)
	io.WriteString(writer, uPath)
	io.WriteString(writer, realIp)
	io.WriteString(writer, header)

}
