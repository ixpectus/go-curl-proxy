package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/ixpectus/fw"
	"github.com/moul/http2curl"
)

var (
	logDirectory = flag.String("d", "./queries", "directory for logs")
	verbose      = flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr         = flag.String("addr", ":9998", "proxy listen address")
)

func main() {
	flag.Parse()
	fmt.Printf("\n>>> startint at port %v <<< debug\n", *addr)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		command, _ := http2curl.GetCurlCommand(req)
		fw.WriteNewFileByMask(*logDirectory, ctx.Req.URL.Hostname()+"-req", []byte(command.String()))
		return req, nil
	})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		fw.WriteNewFileByMask(*logDirectory, ctx.Req.URL.Hostname()+"-resp", bodyBytes)
		resp.Body.Close() //  must close
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		return resp
	})
	proxy.Verbose = *verbose
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
