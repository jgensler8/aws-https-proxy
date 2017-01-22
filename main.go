package main

import (
	"flag"
	"log"
  // "net/url"
	"net/http"
	"regexp"
  "bytes"
  "io/ioutil"

	"github.com/elazarl/goproxy"
  "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func copyHTTPResponse(r *http.Response) *http.Response {
	req := new(http.Response)
	*req = *r

  // Copies Body
  b, err := ioutil.ReadAll(r.Body)
  orPanic(err)
  bb := make([]byte, len(b))
  log.Printf("Len: %v", len(b))
  copy(bb, b)
  r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
  req.Body = ioutil.NopCloser(bytes.NewBuffer(bb))

  // Copies Header
	req.Header = http.Header{}
	for k, v := range r.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	return req
}

func main() {
  sess, err := session.NewSession()
  orPanic(err)

	svc := cloudformation.New(sess)

	params := &cloudformation.DescribeStacksInput{
		NextToken: aws.String("NextToken"),
		StackName: aws.String("StackName"),
	}


	proxy := goproxy.NewProxyHttpServer()
  proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
  proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("amazonaws.com$"))).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    ctx.Logf("%v", "We can see what APIs are being called!")
    return req, ctx.Resp
  })
  proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    ctx.Logf("%v", "We can modify some data coming back!")

    r, out := svc.DescribeStacksRequest(params)
    r.HTTPResponse = copyHTTPResponse(resp)

		r.Handlers.UnmarshalMeta.Run(r)
		r.Handlers.ValidateResponse.Run(r)
		if r.Error != nil {
			r.Handlers.UnmarshalError.Run(r)
			r.Handlers.Retry.Run(r)
			r.Handlers.AfterRetry.Run(r)
			if r.Error != nil {
				ctx.Logf("%v", "Validate Response")
				orPanic(r.Error)
			}
			ctx.Logf("%v", "Validate Response")
		}

		r.Handlers.Unmarshal.Run(r)
		if r.Error != nil {
			r.Handlers.Retry.Run(r)
			r.Handlers.AfterRetry.Run(r)
			if r.Error != nil {
				ctx.Logf("%v", "Unmarshal Response")
				orPanic(r.Error)
			}
			ctx.Logf("%v", "Unmarshal Response")
		}

    ctx.Logf("%v", out.Stacks)

    return resp
  })
	verbose := flag.Bool("v", true, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	proxy.Verbose = *verbose
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
