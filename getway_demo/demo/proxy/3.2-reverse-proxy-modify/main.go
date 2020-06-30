package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

/*
	利用 ReverseProxy 提供的 ModifyResponse 函数对代理访问返回内容进行修改
	访问：127.0.0.1:2002/xxx     返回 127.0.0.1:2003/base/xxx   经过修改返回 test hello 127.0.0.1:2003/base/xxx

	curl http://localhost:2002/pingtest\?sdfs
	test hello http://127.0.0.1:2003/base/pingtest
*/
var Addr = "127.0.0.1:2002"

func main() {
	rs1 := "http://127.0.0.1:2003/base"
	url1, err := url.Parse(rs1)
	if err != nil {
		log.Println(err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(url1)

	// 设置修改函数
	proxy.ModifyResponse = MyselfModifyResp
	// 设置错误处理函数
	proxy.ErrorHandler = MyselfErrorHandler

	err = http.ListenAndServe(Addr, proxy)
	if err != nil {
		log.Println(err.Error())
	}
}

func MyselfModifyResp(resp *http.Response) error {
	// 读取真实服务器返回的数据
	oldPayload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 修改原数据，生成新数据
	newPayload := []byte("test hello " + string(oldPayload))

	// 由于修改了数据，需要重新写入数据
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(newPayload))
	resp.ContentLength = int64(len(newPayload))
	resp.Header.Set("Content-Length", fmt.Sprint(len(newPayload)))

	return nil
}

func MyselfErrorHandler(res http.ResponseWriter, req *http.Request, err error) {
	res.Write([]byte(err.Error()))
}
