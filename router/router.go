package router

import (
	"fmt"
	"github.com/yyxing/glu/context"
	"net/http"
	"strings"
)

var (
	separator = "-"
)

type Router struct {
	roots    map[string]*node
	handlers map[string]context.Handlers
}

func parsePattern(pattern string) []string {
	s := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, part := range s {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return parts
}
func (router *Router) AddRouter(method string, pattern string, handler ...context.Handler) {
	key := method + separator + pattern
	_, ok := router.roots[method]
	if !ok {
		router.roots[method] = &node{children: make([]*node, 0)}
	}
	parts := parsePattern(pattern)
	err := router.roots[method].insert(pattern, parts, 0)
	if err != nil {
		panic(err)
	}
	router.handlers[key] = append(router.handlers[key], handler...)
}

// 查找路由
func (router *Router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := router.roots[method]
	if !ok {
		return nil, nil
	}
	node := root.search(path, searchParts, 0)
	if node == nil {
		return nil, nil
	}
	// 表示找到了path对应的handler
	// 将参数抽取出来
	parts := parsePattern(node.pattern)
	for i, part := range parts {
		_, paramName := pathValid(part)
		if paramName != "" {
			if part[0] == '*' && len(part) > 1 {
				params[paramName] = strings.Join(searchParts[i:], "/")
				break
			}
			params[paramName] = searchParts[i]
		}
	}
	return node, params
}

func (router *Router) Serve(ctx *context.Context) {
	node, params := router.getRoute(ctx.Method, ctx.Path)
	if node != nil {
		ctx.Params = params
		key := ctx.Method + separator + node.pattern
		mainHandler := router.handlers[key]
		ctx.SetHandlers(mainHandler...)
	} else {
		ctx.StatusCode(http.StatusNotFound)
		_, _ = ctx.WriteString(fmt.Sprintf("404 NOT FOUND: %s\n", ctx.Path))
	}
	// 开始触发Handler
	ctx.Next()
}
func NewRouter() *Router {
	return &Router{handlers: make(map[string]context.Handlers), roots: make(map[string]*node)}
}
