package proxy

import (
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

type proxy struct {
	urlMap map[string]proxyFunction
	*proxyDefinitions
}

func newProxy() proxy {
	return proxy{
		urlMap:           make(map[string]proxyFunction),
		proxyDefinitions: &proxyDefinitions{functions: make(map[string]*definition)},
	}
}

type proxyFunction struct {
	fnName    string
	remoteUrl string
}

// single function and its proxyDefinitions definition
type definition struct {
	fnName    string
	proxyFrom string // /functions/fnName
	proxyTo   string // fn.url -> hostname:service-port/entryPoint

	agentAddr string // rpc or http
}

// fn-url and proxyDefinitions
type proxyDefinitions struct {
	functions map[string]*definition
}

const Functions = "/functions"

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

// Pxy https://developpaper.com/how-does-golang-implement-http-proxy-and-reverse-proxy/
type Pxy struct{}

func (p *Pxy) ServeHTTP(rw http.ResponseWriter, req *http.Request, host string) (int, trace.Span) {
	// tr := otel.Tracer(MetricTracerFwToAgent)

	// otel.GetTracerProvider().Tracer(MetricTracerFwToAgent)

	// ctx, span := tr.Start(ctx, HttpProxy)
	// defer span.End()

	transport := otelhttp.NewTransport(http.DefaultTransport)
	//outReq := &http.Request{}
	outReq, _ := http.NewRequestWithContext(req.Context(), req.Method, req.RequestURI, req.Body)
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
	endpoint := strings.Trim(outReq.URL.String(), Functions)
	var err error
	outReq.URL, err = url.Parse(host + "/" + endpoint)
	trace.SpanFromContext(outReq.Context()).SetAttributes(attribute.Key("proxyFrom").String(endpoint))
	trace.SpanFromContext(outReq.Context()).SetAttributes(attribute.Key("proxyTo").String(outReq.URL.String()))
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return http.StatusBadGateway, trace.SpanFromContext(outReq.Context())
	}
	res, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return http.StatusBadGateway, trace.SpanFromContext(outReq.Context())
	}
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
	res.Body.Close()
	return res.StatusCode, trace.SpanFromContext(outReq.Context())
}
