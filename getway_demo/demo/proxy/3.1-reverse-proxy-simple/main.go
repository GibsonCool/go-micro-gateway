package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

/*
	利用官方库提供的  ReverseProxy 实现代理
	访问：127.0.0.1:2002/xxx     返回 127.0.0.1:2003/base/xxx

	curl http://localhost:2002/pingtest\?sdfs
	http://127.0.0.1:2003/base/pingtest
*/
var Addr = "127.0.0.1:2002"

func main() {
	rs1 := "http://127.0.0.1:2003/base"
	url1, err := url.Parse(rs1)
	if err != nil {
		log.Println(err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(url1)
	err = http.ListenAndServe(Addr, proxy)
	if err != nil {
		log.Println(err.Error())
	}
}
