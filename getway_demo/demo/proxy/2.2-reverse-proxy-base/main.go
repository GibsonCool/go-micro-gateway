package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	ProxyAddr = "http://127.0.0.1:2003"
	Port      = "2002"
)

/*
	代理服务器，监听2002端口。
	将请求转发到被代理服务器  2003、2004 端口服务器上
*/
func main() {
	fmt.Println("开启代理服务器，监听端口：", Port)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+Port, nil))
}

/*
	修改要实际访问的目标服务器，然后执行请求拿到结果写入上游，完成反向代理效果
	通过 curl http://localhost:2002/pingtest?sdfs 访问 2002 端口监听服务器
	实际返回的是  http://127.0.0.1:2003/pingtest    real-server 2003 端口监听服务器的处理内容
*/
func handler(writer http.ResponseWriter, request *http.Request) {
	// 解析代理地址，
	parse, err := url.Parse(ProxyAddr)

	// 修改要请求的目标主机为代理地址主机
	request.URL.Scheme = parse.Scheme
	request.URL.Host = parse.Host

	// 下游执行请求
	transport := http.DefaultTransport
	response, err := transport.RoundTrip(request)
	if err != nil {
		log.Print(err)
		return
	}

	// 将下游结果内容返回给上游
	for k, vv := range response.Header {
		for _, v := range vv {
			writer.Header().Add(k, v)
		}
	}
	defer response.Body.Close()
	bufio.NewReader(response.Body).WriteTo(writer)

}
