package glu

import (
	"github.com/yyxing/glu/router"
	"log"
	"net/http"
)

type Engine struct {
	*router.APIBuilder
	router *router.Router
}

func New() *Engine {
	engine := &Engine{APIBuilder: router.NewAPIBuilder()}
	return engine
}

func (e *Engine) Run(addr string) {
	log.Printf("Now listening on: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, e))
}
func (e *Engine) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	e.APIBuilder.HandleRequest(w, request)
}
