package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

/*
	通过两层代理转发访问真实服务器 2.1-real-server ，验证 X-Forwarded-For 和 X-Real-Ip 的安全信息
	127.0.0.1:2001--->127.0.0.1:2002--->127.0.0.1:2003/base

	测试1：
	直接访问	curl http://localhost:2001/hello-test?123
	返回内容：
			一层代理: test hello http://127.0.0.1:2003/base/hello-test
			RemoteAddr=127.0.0.1:57060,X-Forwarded-For=127.0.0.1, 127.0.0.1,X-Real-Ip=127.0.0.1:57058
			headers=map[Accept:[* /*] Accept-Encoding:[gzip] User-Agent:[curl/7.64.1] X-Forwarded-For:[127.0.0.1, 127.0.0.1] X-Real-Ip:[127.0.0.1:57058]]
	可以看出:服务器返回的当前 I P和记录请求过来的 X-Real-Ip 是不同的



	测试2：
	修改X-Forwarded-For：curl -H 'X-Forwarded-For: 2.2.2.2'  http://localhost:2001/hello-test?123
	返回内容：
			一层代理: test hello http://127.0.0.1:2003/base/hello-test
			RemoteAddr=127.0.0.1:57074,X-Forwarded-For=2.2.2.2, 127.0.0.1, 127.0.0.1,X-Real-Ip=127.0.0.1:57072
			headers=map[Accept:[* /*] Accept-Encoding:[gzip] User-Agent:[curl/7.64.1] X-Forwarded-For:[2.2.2.2, 127.0.0.1, 127.0.0.1] X-Real-Ip:[127.0.0.1:57072]]
	可以看出：X-Forwarded-For 的值直接可以通过 curl 的方式伪造



	测试3：
	修改 X-Real-Ip：curl -H 'X-Real-Ip: 3.3.3.3'  http://localhost:2001/hello-test?123
	返回内容：
			一层代理: test hello http://127.0.0.1:2003/base/hello-test
			RemoteAddr=127.0.0.1:57084,X-Forwarded-For=127.0.0.1, 127.0.0.1,X-Real-Ip=127.0.0.1:57082
			headers=map[Accept:[* /*] Accept-Encoding:[gzip] User-Agent:[curl/7.64.1] X-Forwarded-For:[127.0.0.1, 127.0.0.1] X-Real-Ip:[127.0.0.1:57082]]
	可以看出：X-Real-Ip 并没有修改成功，无法伪造
*/

var Addr = "127.0.0.1:2001"

func main() {
	rs1 := "http://127.0.0.1:2002"
	url1, err := url.Parse(rs1)
	if err != nil {
		log.Println(err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(url1)

	// 设置修改函数
	proxy.ModifyResponse = MyselfModifyResp
	// 设置错误处理函数
	proxy.ErrorHandler = MyselfErrorHandler

	// 复制 NewSingleHostReverseProxy 中的方法，从新设置请求协调者
	targetQuery := url1.RawQuery
	proxy.Director = func(req *http.Request) {

		req.URL.Scheme = url1.Scheme
		req.URL.Host = url1.Host
		req.URL.Path = singleJoiningSlash(url1.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		// 修改，增加请求头设置
		req.Header.Set("X-Real-Ip", req.RemoteAddr)
	}

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
	newPayload := []byte("一层代理: " + string(oldPayload))

	// 由于修改了数据，需要重新写入数据
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(newPayload))
	resp.ContentLength = int64(len(newPayload))
	resp.Header.Set("Content-Length", fmt.Sprint(len(newPayload)))

	return nil
}

func MyselfErrorHandler(res http.ResponseWriter, req *http.Request, err error) {
	res.Write([]byte(err.Error()))
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
