package main

import (
	"fmt"
	"go-micro-gateway/getway_demo/proxy/middleware"
	"go-micro-gateway/getway_demo/proxy/proxy"
	"log"
	"net/http"
	"net/url"
)

var addr = "127.0.0.1:2002"

func main() {

	// 反向代理
	reverseProxy := func(c *middleware.SliceRouterContext) http.Handler {
		rs1 := "http://127.0.0.1:2003/base"
		url1, err1 := url.Parse(rs1)
		if err1 != nil {
			log.Println(err1)
		}

		rs2 := "http://127.0.0.1:2004/base"
		url2, err2 := url.Parse(rs2)
		if err2 != nil {
			log.Println(err2)
		}

		urls := []*url.URL{url1, url2}
		return proxy.NewMultipleHostsReverseProxy(c, urls)
	}

	// 初始化路由
	sliceRouter := middleware.NewSliceRouter()

	// /base 路径下的请求 中间件来添加业务逻辑代码
	sliceRouter.Group("/base").Use(TraceLogSliceMw, func(ctx *middleware.SliceRouterContext) {
		ctx.Rw.Write([]byte("--来自中间件插入的的内容--"))
	})

	// /proxy 路径下的请求，通过方向代理处理请求，并屏蔽 coreFunc 的处理
	sliceRouter.Group("/proxy").Use(TraceLogSliceMw, func(ctx *middleware.SliceRouterContext) {
		reverseProxy(ctx).ServeHTTP(ctx.Rw, ctx.Req)
		// 停止后面的中间件调用，间接屏蔽 coreFunc 的逻辑
		ctx.Abort()
	})

	// 最终核心业务处理逻辑
	coreFunc := func(routerContext *middleware.SliceRouterContext) http.Handler {
		return &coreFunc{}
	}

	routerHandler := middleware.NewSliceRouterHandler(coreFunc, sliceRouter)

	http.ListenAndServe(addr, routerHandler)
}

type coreFunc struct {
}

func (c coreFunc) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("来自 coreFunc 的核心处理内容"))
}

// 日志中间件
func TraceLogSliceMw(ctx *middleware.SliceRouterContext) {
	fmt.Println("日志记录 before", ctx.Req.RequestURI)
	ctx.Next()
	fmt.Println("日志记录 after")
}
