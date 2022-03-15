package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Ishan27g/ryo-Faas/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const Functions = "/functions"

type definition struct {
	fnName    string
	proxyFrom string // /functions/fnName
	proxyTo   string // fn.url -> hostname:service-port/entryPoint

	agentAddr string // rpc or http
}
type proxy struct {
	urlMap map[string]proxyFunction
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
		urlMap:           make(map[string]proxyFunction),
		proxyDefinitions: &proxyDefinitions{functions: make(map[string]*definition)},
	}
}

func (p *proxyDefinitions) details() []types.FunctionJsonRsp {
	var str []types.FunctionJsonRsp
	for _, d := range p.functions {
		str = append(str, types.FunctionJsonRsp{
			Name: d.fnName,
			// Status:  "?",
			Url:     d.proxyFrom,
			Proxy:   d.proxyTo,
			AtAgent: d.agentAddr,
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
func (p *proxyDefinitions) getFn(fnName string) *definition {
	if p.functions[fnName] == nil {
		fmt.Println("not found in p.functions", fnName)
		return nil
	}
	return p.functions[fnName]
}
func (p *proxyDefinitions) get(fnName string) (*Pxy, string, string) {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		fmt.Println("not found in proxyDefinitions", fnName)
		return nil, "", ""
	}
	return new(Pxy), p.functions[fnName].proxyTo, p.functions[fnName].agentAddr
}
func (p *proxyDefinitions) urlPair(fnName string) (string, string) {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		return "", ""
	}
	return p.functions[fnName].proxyTo, p.functions[fnName].proxyFrom

}
func (p *proxyDefinitions) add(fn types.FunctionJsonRsp) string {
	d := &definition{
		fnName:    fn.Name,
		proxyFrom: Functions + "/" + strings.ToLower(fn.Name),
		proxyTo:   fn.Proxy,
		agentAddr: fn.AtAgent,
	}
	p.functions[strings.ToLower(fn.Name)] = d
	fmt.Println("ADDED PROXY ", d)
	return d.proxyFrom
}

type Pxy struct{}

func (p *Pxy) ServeHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request, host string) (int, trace.Span) {
	// tr := otel.Tracer(MetricTracerFwToAgent)

	// otel.GetTracerProvider().Tracer(MetricTracerFwToAgent)

	// ctx, span := tr.Start(ctx, HttpProxy)
	// defer span.End()

	transport := otelhttp.NewTransport(http.DefaultTransport)
	//outReq := &http.Request{}
	outReq, _ := http.NewRequestWithContext(ctx, req.Method, req.RequestURI, req.Body)
	//*outReq = *req
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
	endpoint := strings.TrimPrefix(outReq.URL.RequestURI(), Functions)
	fmt.Println("Sending to ", host+endpoint)
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
