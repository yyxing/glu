package main

import (
	"bytes"
	"fmt"
	"github.com/yyxing/glu"
	"github.com/yyxing/glu/context"
	"github.com/yyxing/glu/middleware/limiter"
	"github.com/yyxing/glu/middleware/logger"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var (
	addr = "127.0.0.1:2002"
)

func main() {
	rs1 := "http://127.0.0.1:2003/base"
	url1, err := url.Parse(rs1)
	if err != nil {
		log.Fatal(err)
	}
	rs2 := "http://127.0.0.1:2004/base"
	url2, err := url.Parse(rs2)
	if err != nil {
		log.Fatal(err)
	}
	proxy := NewHostsReverseProxy(url1, url2)
	engine := glu.New()
	l := logger.New()
	v1 := engine.Group("/", limiter.New(1*time.Second, 2), l)
	{
		v1.Get("/:name", func(c *context.Context) {
			c.WriteString("testttt")
		})
	}
	engine.Proxy("/base", proxy)
	engine.Run(addr)
	//log.Println(http.ListenAndServe(addr, proxy))
}
func NewHostsReverseProxy(targets ...*url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		target := targets[rand.Intn(len(targets))]
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	modifyFunc := func(response *http.Response) error {
		if response.StatusCode > 299 && response.StatusCode < 200 {
			// 获取原数据的payload
			oldPayload, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}
			newPayload := []byte("hello " + string(oldPayload))
			nopCloser := ioutil.NopCloser(bytes.NewBuffer(newPayload))
			response.Body = nopCloser
			response.ContentLength = int64(len(newPayload))
			response.Header.Set("Content-Length", fmt.Sprint(len(newPayload)))
		}
		return nil
	}
	return &httputil.ReverseProxy{Director: director, ModifyResponse: modifyFunc}
}
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}
