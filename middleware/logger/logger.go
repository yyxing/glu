package logger

import (
	"github.com/yyxing/glu/context"
	"log"
	"time"
)

func New() context.Handler {
	return func(c *context.Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v for group v2", 200, c.Request.RequestURI, time.Since(t))
	}
}
