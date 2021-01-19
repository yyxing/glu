package logger

import (
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"github.com/yyxing/glu/context"
	"time"
)

func init() {
	formatter := prefixed.TextFormatter{
		ForceColors:     true,
		DisableColors:   false,
		ForceFormatting: true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:03.00000",
	}
	formatter.SetColorScheme(&prefixed.ColorScheme{
		TimestampStyle: "37",
	})
	log.SetFormatter(&formatter)
}

func New() context.Handler {
	return func(c *context.Context) {
		t := time.Now()
		c.Next()
		log.Infof("[%d] %s in %v", 200, c.Request.RequestURI, time.Since(t))
	}
}
