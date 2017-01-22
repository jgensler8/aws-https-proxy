package easy

import (
	"flag"
	"log"
	"net/http"
	"regexp"

	"github.com/elazarl/goproxy"
)

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
  proxy := goproxy.NewProxyHttpServer()
  proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
  proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("amazonaws.com$"))).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    ctx.Logf("%v", "We can see what APIs are being called!")
    return req, ctx.Resp
  })
  proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    ctx.Logf("%v", "We can modify some data coming back!")
    return resp
  })
  verbose := flag.Bool("v", true, "should every proxy request be logged to stdout")
  proxy.Verbose = *verbose
  log.Fatal(http.ListenAndServe(":8080", proxy))
}
