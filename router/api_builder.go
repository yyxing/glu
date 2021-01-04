package router

import (
	"github.com/yyxing/glu/context"
	"net/http"
)

type APIBuilder struct {
	middlewares context.Handlers
	prefix      string
	router      *Router
}

func NewAPIBuilder() *APIBuilder {
	api := &APIBuilder{
		middlewares: make(context.Handlers, 0),
		prefix:      "/",
		router:      NewRouter(),
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
func (api *APIBuilder) Group(prefix string, handlers ...context.Handler) Group {
	middlewares := joinHandlers(api.middlewares, handlers)
	if api.prefix[len(api.prefix)-1] == '/' && prefix[0] == '/' {
		prefix = prefix[1:]
	}
	if api.prefix[len(api.prefix)-1] != '/' && prefix[0] != '/' {
		prefix = "/" + prefix
	}
	prefix = api.prefix + prefix
	return &APIBuilder{
		middlewares: middlewares,
		prefix:      prefix,
		router:      api.router,
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
	ctx := context.NewContext(w, request)
	//ctx.SetHandlers(api.middlewares...)
	api.router.Serve(ctx)
}
