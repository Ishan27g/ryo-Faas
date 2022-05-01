package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Ishan27g/ryo-Faas/pkg/metric"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const Functions = "/functions"

type definition struct {
	fnName    string
	proxyFrom string // /functions/fnName
	// proxyTo   string // fn.url -> hostname:service-port/entryPoint

	isAsync  bool
	isMain   bool
	instance int
	proxyTo  string
}
type balancer struct {
	groups map[string]*balancerGroup
}

func (b *balancer) Add(instance int, fnName, url string) {
	if b.groups[fnName] == nil {
		b.groups[fnName] = &balancerGroup{
			lock:      sync.Mutex{},
			fnName:    fnName,
			urls:      map[int]*inst{},
			currIndex: 0,
		}
	}
	b.groups[fnName].Add(instance, url)
	fmt.Println("balancer - added upstream ", url, " for ", fnName)
	fmt.Println("balancer - all upstreams ")
	for fnName, bg := range b.groups {
		fmt.Println("\t", fnName)
		for _, s := range bg.urls {
			fmt.Println(*s)
		}
	}
}
func (b *balancer) GetNext(fnName string) string {
	return b.groups[fnName].GetNext()
}

type inst struct {
	url      *string
	instance int
}
type balancerGroup struct {
	lock      sync.Mutex
	fnName    string
	urls      map[int]*inst
	currIndex int
}

func (b *balancerGroup) Add(instance int, url string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	fmt.Println("len before add  is ", len(b.urls))
	b.urls[len(b.urls)] = &inst{
		url:      &url,
		instance: instance,
	}
	fmt.Println("len afterwards is ", len(b.urls))

}
func (b *balancerGroup) Remove(url string) int {
	fmt.Println("--------------------------------------")
	defer fmt.Println("--------------------------------------")
	b.lock.Lock()
	defer b.lock.Unlock()
	var existing []*inst
	var removedInst int
	fmt.Println("len before remove  is ", len(b.urls))
	fmt.Println("removing ", url)
	for i, s := range b.urls {
		if *s.url == url {
			removedInst = s.instance
		} else {
			fmt.Println("no matched ", *s.url)
			existing = append(existing, s)
		}
		delete(b.urls, i)
	}
	b.currIndex = 0
	for i, s := range existing {
		b.urls[i] = s
		// b.currIndex++
	}
	fmt.Println("len afterwards is ", len(b.urls))
	return removedInst
}
func (b *balancerGroup) GetNext() string {
	b.lock.Lock()
	defer b.lock.Unlock()
	service := b.urls[b.currIndex]
	b.currIndex = (b.currIndex + 1) % len(b.urls)
	return *service.url
}

type proxy struct {
	functions map[string]*definition
	metrics   metric.Monitor
	balancer
}

func newProxy() proxy {
	return proxy{
		functions: make(map[string]*definition),
		metrics:   metric.Start(),
		balancer:  balancer{groups: map[string]*balancerGroup{}},
	}
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

func (p *proxy) details() []types.FunctionJsonRsp {
	var str []types.FunctionJsonRsp
	for _, d := range p.functions {
		str = append(str, types.FunctionJsonRsp{
			Name:    d.fnName,
			Url:     d.proxyFrom,
			Proxy:   d.proxyTo,
			IsAsync: d.isAsync,
			IsMain:  d.isMain,
		})
	}
	for s, bg := range p.groups {
		fmt.Println(s)
		for _, i := range bg.urls {
			fmt.Println(bg.fnName, *i.url, i.instance)
		}
	}
	return str
}
func (p *proxy) remove(fnName string) int {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		return -1
	}
	defer func() {
		if len(p.groups[fnName].urls) == 0 {
			defer delete(p.functions, fnName)
			defer delete(p.groups, fnName)
		}
	}()

	toRemove := p.groups[fnName].GetNext()
	fmt.Println("p.groups[fnName].GetNext() -> toRemove -> ", toRemove)
	return p.groups[fnName].Remove(toRemove)
}
func (p *proxy) asDefinition(fnName string) *definition {
	if p.functions[fnName] == nil {
		fmt.Println("not found in p.functions", fnName)
		return nil
	}
	return p.functions[fnName]
}
func (p *proxy) getFuncFwHost(fnName string) string {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		fmt.Println("not found in proxyDefinitions", fnName)
		return ""
	}
	return p.getUpstreamFor(fnName)
}
func (p *proxy) invoke(fnName string) (*Pxy, string, bool, bool) {
	fnName = strings.ToLower(fnName)
	if p.functions[fnName] == nil {
		fmt.Println("not found in proxyDefinitions", fnName)
		return nil, "", false, false
	}
	p.metrics.Invoked(fnName)
	upstream := p.getUpstreamFor(fnName)
	return new(Pxy), upstream, p.functions[fnName].isAsync, p.functions[fnName].isMain
}

func (p *proxy) getUpstreamFor(fnName string) string {
	return p.balancer.GetNext(fnName)
}
func (p *proxy) add(fn types.FunctionJsonRsp, instance int) string {
	fmt.Println("Adding PROXY ", fn)
	d := &definition{
		fnName:    fn.Name,
		proxyFrom: Functions + "/" + strings.ToLower(fn.Name),
		proxyTo:   fn.Proxy,
		isAsync:   fn.IsAsync,
		isMain:    fn.IsMain,
		instance:  instance,
	}
	p.functions[strings.ToLower(fn.Name)] = d
	p.balancer.Add(instance, strings.ToLower(fn.Name), d.proxyTo)
	metric.Register(fn.Name)

	fmt.Println("ADDED PROXY ", p.functions[strings.ToLower(fn.Name)])
	return d.proxyFrom
}

type Pxy struct{}

func (p *Pxy) ServeHTTP(rw http.ResponseWriter, outReq *http.Request, host string, trimServiceName string) (int, trace.Span) {

	transport := otelhttp.NewTransport(http.DefaultTransport)
	endpoint := ""
	if trimServiceName != "" {
		endpoint = strings.TrimPrefix(outReq.URL.RequestURI(), Functions+"/"+trimServiceName)
	} else {
		endpoint = strings.TrimPrefix(outReq.URL.RequestURI(), Functions)
	}

	fmt.Println("Forwarding to ", host+endpoint)
	var err error
	outReq.URL, err = url.Parse(host + endpoint)

	span := trace.SpanFromContext(outReq.Context())
	span.SetAttributes(semconv.HTTPTargetKey.String(outReq.URL.String()))

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
