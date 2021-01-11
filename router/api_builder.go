package router

import (
	"github.com/yyxing/glu/context"
	"net/http"
	"sync"
)

type APIBuilder struct {
	middlewares  context.Handlers
	prefix       string
	router       *Router
	proxyHandler http.Handler
	pool         sync.Pool
}

func NewAPIBuilder() *APIBuilder {
	api := &APIBuilder{
		middlewares: make(context.Handlers, 0),
		prefix:      "/",
		router:      NewRouter(),
		pool: sync.Pool{New: func() interface{} {
			return context.NewContext()
		}},
	}
	return api
}
func (api *APIBuilder) addRoute(method string, pattern string, handler context.Handler) {
	if api.prefix[len(api.prefix)-1] == '/' {
		if pattern[0] == '/' {
			pattern = pattern[1:]
		}
	}
	pattern = api.prefix + pattern
	handlers := append(api.middlewares, handler)
	api.router.AddRouter(method, pattern, handlers...)
}
func joinHandlers(h1 context.Handlers, h2 context.Handlers) context.Handlers {
	nowLen := len(h1)
	newLen := nowLen + len(h2)
	newHandlers := make(context.Handlers, newLen)
	copy(newHandlers, h1)
	copy(newHandlers[nowLen:], h2)
	return newHandlers
}
func jsonPrefixPath(a, b string) string {
	if a[len(a)-1] == '/' && b[0] == '/' {
		b = b[1:]
	}
	if a[len(a)-1] != '/' && b[0] != '/' {
		b = "/" + b
	}
	b = a + b
	return b
}
func (api *APIBuilder) Group(prefix string, handlers ...context.Handler) Group {
	middlewares := joinHandlers(api.middlewares, handlers)
	prefix = jsonPrefixPath(api.prefix, prefix)
	return &APIBuilder{
		middlewares: middlewares,
		prefix:      prefix,
		router:      api.router,
	}
}
func (api *APIBuilder) ReverseProxy(prefix string, handler http.Handler) Group {
	prefix = jsonPrefixPath(api.prefix, prefix)
	return &APIBuilder{
		middlewares:  api.middlewares,
		prefix:       prefix,
		router:       api.router,
		proxyHandler: handler,
	}
}
func (api *APIBuilder) Use(handler ...context.Handler) {
	api.middlewares = append(api.middlewares, handler...)
}
func (api *APIBuilder) Get(pattern string, handler context.Handler) {
	api.addRoute(http.MethodGet, pattern, handler)
}

func (api *APIBuilder) Head(pattern string, handler context.Handler) {
	api.addRoute(http.MethodHead, pattern, handler)
}

func (api *APIBuilder) Delete(pattern string, handler context.Handler) {
	api.addRoute(http.MethodDelete, pattern, handler)
}

func (api *APIBuilder) Post(pattern string, handler context.Handler) {
	api.addRoute(http.MethodPost, pattern, handler)
}

func (api *APIBuilder) Options(pattern string, handler context.Handler) {
	api.addRoute(http.MethodOptions, pattern, handler)
}

func (api *APIBuilder) Put(pattern string, handler context.Handler) {
	api.addRoute(http.MethodPut, pattern, handler)
}

func (api *APIBuilder) Patch(pattern string, handler context.Handler) {
	api.addRoute(http.MethodPatch, pattern, handler)
}

func (api *APIBuilder) Trace(pattern string, handler context.Handler) {
	api.addRoute(http.MethodTrace, pattern, handler)
}

func (api *APIBuilder) Connect(pattern string, handler context.Handler) {
	api.addRoute(http.MethodConnect, pattern, handler)
}

func (api *APIBuilder) Handle(method string, pattern string, handler context.Handler) {
	api.addRoute(method, pattern, handler)
}

func (api *APIBuilder) HandleRequest(w http.ResponseWriter, request *http.Request) {
	ctx := api.pool.Get().(*context.Context)
	ctx.Request = request
	ctx.Writer = w
	ctx.Reset()
	api.router.Serve(ctx)
}

func (api *APIBuilder) Prefix() string {
	return api.prefix
}

func (api *APIBuilder) ProxyHandler() http.Handler {
	return api.proxyHandler
}
