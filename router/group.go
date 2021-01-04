package router

import "github.com/yyxing/glu/context"

// 分组路由
type Group interface {
	// HTTP 请求
	Get(pattern string, handler context.Handler)
	Post(pattern string, handler context.Handler)
	Put(pattern string, handler context.Handler)
	Patch(pattern string, handler context.Handler)
	Head(pattern string, handler context.Handler)
	Connect(pattern string, handler context.Handler)
	Delete(pattern string, handler context.Handler)
	Options(pattern string, handler context.Handler)
	Trace(pattern string, handler context.Handler)
	// 添加路由信息
	Handle(method string, pattern string, handler context.Handler)
	// 创建分组
	Group(prefix string, handlers ...context.Handler) Group
	// 中间件注入
	Use(handler ...context.Handler)
}
