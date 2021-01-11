package limiter

import (
	"github.com/yyxing/glu/context"
	"golang.org/x/time/rate"
	"time"
)

func New(limit time.Duration, maximum int) context.Handler {
	limiter := rate.NewLimiter(rate.Every(limit), maximum)
	return func(c *context.Context) {
		if !limiter.Allow() {
			_, _ = c.WriteString("请求过快，请稍后再试！")
			c.Abort()
			return
		}
		c.Next()
	}
}
