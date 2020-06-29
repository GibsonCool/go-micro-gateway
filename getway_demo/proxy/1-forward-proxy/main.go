package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Pxy struct {
}

func (p Pxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Printf("接收到请求 %s %s %s \n", request.Method, request.Host, request.RemoteAddr)
	// 1.浅拷贝对象，然后在此基础上增加本机真实IP
	outReq := new(http.Request)
	*outReq = *request

	if clientIp, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		fmt.Println("clientIp：", clientIp)
		if oldIpList, ok := outReq.Header["X-Forwarded-For"]; ok {

			oldIpList = append(oldIpList, clientIp)
			clientIp = strings.Join(oldIpList, ",")
			fmt.Println("clientIp after add：", clientIp)
		}
		outReq.Header.Set("X-Forwarded-For", clientIp)
	}

	// 2.请求下游
	response, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadGateway)
		return
	}

	// 3.将下游内容返回给上游
	for key, value := range response.Header {
		for _, v := range value {
			writer.Header().Add(key, v)
		}
	}
	writer.WriteHeader(response.StatusCode)
	io.Copy(writer, response.Body)
	response.Body.Close()
}

/*
	浏览器代理
*/
func main() {
	fmt.Println("监听 8080 端口")
	http.Handle("/", &Pxy{})
	http.ListenAndServe("0.0.0.0:8080", nil)
}
