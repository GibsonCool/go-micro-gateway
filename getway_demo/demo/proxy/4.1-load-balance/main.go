package main

import (
	"bytes"
	"fmt"
	load_balance "go-micro-gateway/getway_demo/proxy/load-balance"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var (
	addr      = "127.0.0.1:2002"
	transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, //连接超时
			KeepAlive: 30 * time.Second, //长连接超时时间
		}).DialContext,
		MaxIdleConns:          100,              //最大空闲连接
		IdleConnTimeout:       90 * time.Second, //空闲超时时间
		TLSHandshakeTimeout:   10 * time.Second, //tls握手超时时间
		ExpectContinueTimeout: 1 * time.Second,  //100-continue状态码超时时间
	}
)

func NewMultipleHostReverseProxy(lb load_balance.LoadBalance) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		// 根据不同取 key 的方式也交由不同均衡策略获取目标服务器地址
		key := req.RemoteAddr

		targetUrl, err := lb.Get(key)
		if err != nil {
			fmt.Println("通过负载均衡策略获取下一次目标服务器地址失败：", err.Error())
		}
		fmt.Println("key :", key, "target url :", targetUrl)
		target, err := url.Parse(targetUrl)
		if err != nil {
			fmt.Println("目标服务器地址解析失败：", err.Error())
		}
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	modifyFunc := func(resp *http.Response) error {
		// 读取真实服务器返回的数据
		oldPayload, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// 修改原数据，生成新数据
		newPayload := []byte("负载均衡层: " + string(oldPayload))

		// 由于修改了数据，需要重新写入数据
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(newPayload))
		resp.ContentLength = int64(len(newPayload))
		resp.Header.Set("Content-Length", fmt.Sprint(len(newPayload)))

		return nil
	}

	errFunc := func(res http.ResponseWriter, req *http.Request, err error) {
		//res.Write([]byte(err.Error()))
		http.Error(res, "负载均衡层发生错误，ErrorHandler err:"+err.Error(), 500)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyFunc,
		ErrorHandler:   errFunc,
	}
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

/*
	通过传入不同参数类型，获取不同负载均衡策略

	通过修改 ReverseProxy 中 director 中获取 key 的不同方式交由不同负载均衡侧率获取命中目标服务器地址进行转发请求
*/
func main() {
	lb := load_balance.LoadBalanceFactory(load_balance.LbWeightRoundRobin)
	// 添加两条
	if err := lb.Add("http://127.0.0.1:2003", "10"); err != nil {
		fmt.Println(err)
	}

	if err := lb.Add("http://127.0.0.1:2004", "20"); err != nil {
		fmt.Println(err)
	}

	proxy := NewMultipleHostReverseProxy(lb)

	fmt.Println("服务器开启，", addr)
	http.ListenAndServe(addr, proxy)
}
