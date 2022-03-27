package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const Functions = "/functions"

type definition struct {
	fnName    string
	proxyFrom string // /functions/fnName
	proxyTo   string // fn.url -> hostname:service-port/entryPoint

	agentAddr   string // rpc or http
	isAsyncNats bool
	isMain      bool
}
type proxy struct {
	*proxyDefinitions
}

type proxyDefinitions struct {
	functions map[string]*definition
}

type proxyFunction struct {
	fnName    string
	remoteUrl string
}

func newProxy() proxy {
	return proxy{
		proxyDefinitions: &proxyDefinitions{functions: make(map[string]*definition)},
	}
}

func (p *proxyDefinitions) details() []types.FunctionJsonRsp {
	var str []types.FunctionJsonRsp
	for _, d := range p.functions {
		str = append(str, types.FunctionJsonRsp{
			Name:    d.fnName,
			Url:     d.proxyFrom,
			Proxy:   d.proxyTo,
			AtAgent: d.agentAddr,
			IsAsync: d.isAsyncNats,
		})
	}
	return str
}
func (p *proxyDefinitions) remove(fnName string) {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		return
	}
	delete(p.functions, fnName)
}
func (p *proxyDefinitions) asDefinition(fnName string) *definition {
	if p.functions[fnName] == nil {
		fmt.Println("not found in p.functions", fnName)
		return nil
	}
	return p.functions[fnName]
}
func (p *proxyDefinitions) getFuncFwHost(fnName string) string {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		fmt.Println("not found in proxyDefinitions", fnName)
		return ""
	}
	return p.functions[fnName].proxyTo
}
func (p *proxyDefinitions) get(fnName string) (*Pxy, string, string, bool, bool) {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		fmt.Println("not found in proxyDefinitions", fnName)
		return nil, "", "", false, false
	}
	return new(Pxy), p.functions[fnName].proxyTo, p.functions[fnName].agentAddr, p.functions[fnName].isAsyncNats, p.functions[fnName].isMain
}
func (p *proxyDefinitions) add(fn types.FunctionJsonRsp) string {
	d := &definition{
		fnName:      fn.Name,
		proxyFrom:   Functions + "/" + strings.ToLower(fn.Name),
		proxyTo:     fn.Proxy,
		agentAddr:   fn.AtAgent,
		isAsyncNats: fn.IsAsync,
		isMain:      fn.IsMain,
	}
	p.functions[strings.ToLower(fn.Name)] = d
	fmt.Println("ADDED PROXY ", d)
	return d.proxyFrom
}

type Pxy struct{}

func (p *Pxy) ServeHTTP(ctx context.Context, rw http.ResponseWriter, outReq *http.Request, host string, trimServiceName string) (int, trace.Span) {

	transport := otelhttp.NewTransport(http.DefaultTransport)
	//outReq := &http.Request{}
	//*outReq = *req
	// outReq := newRequestWithCtx(ctx, req)
	endpoint := ""
	if trimServiceName != "" {
		endpoint = strings.TrimPrefix(outReq.URL.RequestURI(), Functions+"/"+trimServiceName)
	} else {
		endpoint = strings.TrimPrefix(outReq.URL.RequestURI(), Functions)
	}

	fmt.Println("Fordaring to ", host+endpoint)
	var err error
	outReq.URL, err = url.Parse(host + endpoint)

	span := trace.SpanFromContext(outReq.Context())

	span.SetAttributes(attribute.Key("proxyFrom").String(endpoint))
	span.SetAttributes(attribute.Key("proxyTo").String(outReq.URL.String()))
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return http.StatusBadGateway, span
	}
	res, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return http.StatusBadGateway, span
	}
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
	res.Body.Close()
	return res.StatusCode, span
}

func newFwRequestWithCtx(ctx context.Context, req *http.Request) *http.Request {
	outReq, _ := http.NewRequestWithContext(ctx, req.Method, req.RequestURI, req.Body)

	for key, value := range req.Header {
		for _, v := range value {
			outReq.Header.Add(key, v)
		}
	}
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}
	return outReq
}
