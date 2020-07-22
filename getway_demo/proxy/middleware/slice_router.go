package middleware

import (
	"context"
	"math"
	"net/http"
	"strings"
)

const abortIndex int8 = math.MaxInt8 / 2 //最多 63 个中间件

type SliceGroup struct {
	*SliceRouter
	path     string
	handlers []HandlerFunc
}

type SliceRouter struct {
	groups []*SliceGroup
}

func NewSliceRouter() *SliceRouter {
	return &SliceRouter{}
}

func (r *SliceRouter) Group(path string) *SliceGroup {
	return &SliceGroup{
		SliceRouter: r,
		path:        path,
	}
}
func (g *SliceGroup) Use(meddlers ...HandlerFunc) *SliceGroup {
	g.handlers = append(g.handlers, meddlers...)

	// 判断这个路由组是否已经存在，防止重复添加
	existsFlag := false
	for _, oldGroup := range g.SliceRouter.groups {
		if oldGroup == g {
			existsFlag = true
			break
		}
	}
	if !existsFlag {
		g.SliceRouter.groups = append(g.SliceRouter.groups, g)
	}
	return g
}

type SliceRouterContext struct {
	Rw  http.ResponseWriter
	Req *http.Request
	Ctx context.Context
	*SliceGroup
	index int8
}

func (c *SliceRouterContext) Get(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

func (c *SliceRouterContext) Set(key, val interface{}) {
	c.Ctx = context.WithValue(c.Ctx, key, val)
}

func (c *SliceRouterContext) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *SliceRouterContext) Abort() {
	c.index = abortIndex
}

func (c *SliceRouterContext) IsAborted() bool {
	return c.index >= abortIndex
}
func (c *SliceRouterContext) Reset() {
	c.index = -1
}

func newSliceRouterContext(rw http.ResponseWriter, req *http.Request, r *SliceRouter) *SliceRouterContext {
	newSliceGroup := &SliceGroup{}

	// 最长URL前缀匹配
	matchUrlLen := 0
	for _, group := range r.groups {
		//fmt.Println("req.RequestURI", req.RequestURI)
		//fmt.Println("group.path", group.path)

		// 遍历找出符合 path 前缀的路由组
		if strings.HasPrefix(req.RequestURI, group.path) {
			// 寻找最长符合前缀的，比如 uri: /test/11/22   path1:/test  path2:/test/11 使用 path2
			pathLen := len(group.path)
			if pathLen > matchUrlLen {
				matchUrlLen = pathLen
				*newSliceGroup = *group
			}
		}

	}
	c := &SliceRouterContext{Rw: rw, Req: req, Ctx: req.Context(), SliceGroup: newSliceGroup}
	c.Reset()
	return c
}

type HandlerFunc func(*SliceRouterContext)

type SliceRouterHandler struct {
	coreFunc func(routerContext *SliceRouterContext) http.Handler
	router   *SliceRouter
}

func NewSliceRouterHandler(coreFunc func(routerContext *SliceRouterContext) http.Handler, router *SliceRouter) *SliceRouterHandler {
	return &SliceRouterHandler{coreFunc: coreFunc, router: router}
}

func (s *SliceRouterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := newSliceRouterContext(writer, request, s.router)

	if s.router != nil {
		c.handlers = append(c.handlers, func(c *SliceRouterContext) {
			s.coreFunc(c).ServeHTTP(writer, request)
		})
	}
	c.Reset()
	c.Next()
}
