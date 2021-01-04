package main

import (
	"fmt"
	"github.com/yyxing/glu"
	"github.com/yyxing/glu/context"
	"github.com/yyxing/glu/middleware/gluRecover"
	"github.com/yyxing/glu/middleware/logger"
	"net/http"
)

type H struct {
}

func main() {
	r := glu.New()
	r.Get("/index", func(c *context.Context) {
		c.StatusCode(http.StatusOK)
		c.HTML("<h1>Hello Gee</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.Get("/", func(c *context.Context) {
			c.StatusCode(http.StatusOK)
			c.HTML("<h1>Hello Gee</h1>")
		})
		v1.Use(logger.New())
		v1.Use(gluRecover.New())
		v1.Get("/hello", func(c *context.Context) {
			c.StatusCode(http.StatusOK)
			// expect /hello?name=geektutu
			c.WriteString(fmt.Sprintf("hello %s, you're at %s\n", c.Query("name"), c.Path))
		})
		v1.Get("/panic", func(c *context.Context) {
			panic("test panic")
		})
	}
	v2 := r.Group("/v2")
	{
		v2.Get("/hello/:name", func(c *context.Context) {
			// expect /hello/geektutu
			// expect /hello?name=geektutu
			c.WriteString(fmt.Sprintf("hello %s, you're at %s\n", c.Param("name"), c.Path))
		})

	}
	r.Run(":9999")
}
