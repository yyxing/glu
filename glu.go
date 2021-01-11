package glu

import (
	"github.com/yyxing/glu/middleware/gluRecover"
	"github.com/yyxing/glu/middleware/logger"
	"github.com/yyxing/glu/router"
	"log"
	"net/http"
	"strings"
)

type Engine struct {
	*router.APIBuilder
	proxyGroups []router.Group
}

func New() *Engine {
	engine := &Engine{APIBuilder: router.NewAPIBuilder()}
	return engine
}

func Default() *Engine {
	engine := &Engine{APIBuilder: router.NewAPIBuilder()}
	engine.Use(logger.New(), gluRecover.New())
	return engine
}
func (e *Engine) Proxy(prefix string, handler http.Handler) {
	proxyGroup := e.ReverseProxy(prefix, handler)
	e.proxyGroups = append(e.proxyGroups, proxyGroup)
}
func (e *Engine) Run(addr string) {
	log.Printf("Now listening on: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, e))
}
func (e *Engine) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	for _, group := range e.proxyGroups {
		if strings.HasPrefix(request.URL.Path, group.Prefix()) {
			group.ProxyHandler().ServeHTTP(w, request)
			return
		}
	}
	e.APIBuilder.HandleRequest(w, request)
}
