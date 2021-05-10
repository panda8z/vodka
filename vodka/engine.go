package vodka

import (
	"fmt"
	"net/http"
)

// HandlerFunc 定义了 Engine 自己的处理器。
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 定义了 Vodka 框架的引擎。
type Engine struct {
	router map[string]HandlerFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (e *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	e.router[key] = handler
}

func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter("POST", pattern, handler)
}

// Run 调用**标准库**的 `http.ListenAndServe` 方法启动服务器程序,
// 因为 **Engine** 实现了 `http.Handler` 接口 所以这里的第二个参数可以直接传 **Engine**。
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

// ServeHTTP 实现了 `http.Handler` 接口。
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
