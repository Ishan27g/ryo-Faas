package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/pkg/docker"
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerFwToAgent = "proxy-function-call"
	ServiceName     = "rfa-proxy"

	DefaultRpc           = ":9998"
	DefaultHttp          = ":9999"
	DeployedFunctionPort = ":6000"
)

type handler struct {
	g               *gin.Engine
	httpFnProxyPort string

	proxies proxy

	tracing.MetricManager

	*log.Logger
}

func prettyJson(js interface{}) string {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(data)
}
func (h *handler) Deploy(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {

	span := trace.SpanFromContext(ctx)

	response := new(deploy.DeployResponse)
	var buildHostName = func(entrypoint string) string {
		hostname, _ := os.Hostname()
		if strings.Contains(hostname, "Ishan") { // if running locally
			return "http://" + "localhost"
		}
		// if in docker network
		return "http://" + "rfa-deploy-" + strings.ToLower(request.Functions[0].Entrypoint)
	}
	hnFn := buildHostName(request.Functions[0].Entrypoint) + DeployedFunctionPort
	for _, function := range request.Functions {
		function.ProxyServiceAddr = hnFn
		jsonFn := types.RpcFunctionRspToJson(function)

		proxyUrl := h.proxies.add(jsonFn)
		function.Url = "http://localhost" + h.httpFnProxyPort + proxyUrl
		response.Functions = append(response.Functions, function)

		h.MetricManager.Register(function)

		addFunctionAttributes(&span, function)
	}

	h.Println("DEPLOY RESPONSE IS", prettyJson(response))

	span.AddEvent(prettyJson(response))
	return response, nil
}

func addFunctionAttributes(span *trace.Span, function *deploy.Function) {
	(*span).SetAttributes(attribute.Key(tracing.Entrypoint).String(function.Entrypoint))
	(*span).SetAttributes(attribute.Key(tracing.Url).String(function.Url))
	(*span).SetAttributes(attribute.Key(tracing.Status).String(function.Status))
	(*span).SetAttributes(attribute.Key(tracing.IsMain).Bool(function.IsMain))
	(*span).SetAttributes(attribute.Key(tracing.IsAsync).Bool(function.Async))
}

func (h *handler) Stop(ctx context.Context, request *deploy.Empty) (*deploy.DeployResponse, error) {
	response := new(deploy.DeployResponse)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Key(request.GetEntrypoint()).String(request.GetEntrypoint()))

	if docker.New().StopFunction(strings.ToLower(request.GetEntrypoint())) != nil {
		h.Println("Unable to stop ", request.GetEntrypoint())
		span.AddEvent("Unable to stop " + request.GetEntrypoint())
		return response, nil
	}
	h.proxies.remove(request.GetEntrypoint())
	span.AddEvent("Stopped " + request.GetEntrypoint())
	return response, nil
}

func (h *handler) Details(ctx context.Context, _ *deploy.Empty) (*deploy.DeployResponse, error) {

	span := trace.SpanFromContext(ctx)
	var details = new(deploy.DeployResponse)
	for _, rsp := range h.proxies.details() {
		df := &deploy.Function{
			Entrypoint:       rsp.Name,
			ProxyServiceAddr: rsp.Proxy,
			Url:              "http://localhost" + h.httpFnProxyPort + rsp.Url,
			Status:           rsp.Status,
			Async:            rsp.IsAsync,
			IsMain:           rsp.IsMain,
		}

		details.Functions = append(details.Functions, df)

		addFunctionAttributes(&span, df)
	}
	h.Println("Proxy details returning -> ", h.proxies.details())
	return details, nil
}

func (h *handler) Upload(deploy.Deploy_UploadServer) error {
	return errors.New("no upload method")
}
func (h *handler) returnMetric(err error, result chan<- bool) {

}
func (h *handler) ForwardToAgentHttp(c *gin.Context) {
	fnName := c.Param("entrypoint")
	var proxyError error = nil
	var statusCode = http.StatusBadGateway

	sp := trace.SpanFromContext(c.Request.Context())

	var ctxR context.Context
	if !sp.IsRecording() {
		ctxR, sp = otel.Tracer(TracerFwToAgent).Start(c.Request.Context(), TracerFwToAgent+"-"+fnName)
	} else {
		ctxR = trace.ContextWithSpan(c.Request.Context(), sp)
	}
	sp.SetAttributes(attribute.Key("Forward-Function-Call").String(fnName))

	newReq := newFwRequestWithCtx(ctxR, c.Request)
	now := time.Now()

	result := h.MetricManager.Invoked(fnName)
	defer func() {
		defer close(result)
		result <- proxyError == nil
	}()

	if proxy, fnServiceHost, isAsyncNats, isMain := h.proxies.get(fnName); fnServiceHost != "" {
		proxyError = nil
		if isAsyncNats {
			FuncFw.NewAsyncNats(fnName, "").HandleAsyncNats(c.Writer, newReq)
			statusCode = http.StatusOK
			sp = updateSpan(sp, "async-nats", statusCode, now, fnName)
		} else {
			var stc int
			var span trace.Span
			if isMain {
				stc, span = proxy.ServeHTTP(c.Writer, newReq, fnServiceHost, strings.ToLower(fnName))
			} else {
				stc, span = proxy.ServeHTTP(c.Writer, newReq, fnServiceHost, "")
			}
			statusCode = stc
			span = updateSpan(span, "http", statusCode, now, fnName)
		}
	} else {
		proxyError = errors.New(fnName + " not found")
		sp.SetAttributes(attribute.Key("No Proxy found").String(fnName))
		c.String(http.StatusBadGateway, "Not found - "+fnName)
	}
}

func updateSpan(sp trace.Span, deploymentType string, statusCode int, now time.Time, fnName string) trace.Span {
	sp.SetAttributes(attribute.Key("deployment").String(deploymentType))
	sp.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
	sp.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
	sp.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(statusCode))))
	return sp
}

func (h *handler) reset(c *gin.Context) {
	h.proxies = newProxy()
	c.Status(http.StatusAccepted)
	h.Println("reset")
}
func (h *handler) metrics(c *gin.Context) {
	span := trace.SpanFromContext(c)
	m := h.MetricManager.GetAll()
	fmt.Println(m)
	span.AddEvent(prettyJson(m))
	c.JSON(http.StatusOK, m)
}

func (h *handler) DetailsHttp(c *gin.Context) {

	span := trace.SpanFromContext(c)

	var details []types.FunctionJsonRsp

	for _, fn := range h.proxies.details() {
		if checkHealth(h.proxies.getFuncFwHost(fn.Name)) {
			pFn := h.proxies.asDefinition(fn.Name)
			details = append(details, types.FunctionJsonRsp{
				Name:    pFn.fnName,
				Proxy:   pFn.proxyTo,
				IsAsync: pFn.isAsync,
				IsMain:  pFn.isMain,
			})

			span.SetAttributes(attribute.Key(fn.Name).String(prettyJson(pFn)))

		} else {
			h.Println(h.proxies.getFuncFwHost(fn.Name) + " unreachable")
			span.SetAttributes(attribute.Key(fn.Name).String("unreachable"))
		}
	}
	span.AddEvent(prettyJson(details))

	c.JSON(200, details)
}

func checkHealth(addr string) bool {
	resp, err := http.Get(addr + "/healthcheck")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func Start(ctx context.Context, grpcPort, http string) {
	h := new(handler)
	h.httpFnProxyPort = http
	h.proxies = newProxy()
	h.MetricManager = tracing.Manager()
	h.Logger = log.New(os.Stdout, ServiceName, log.Ltime)

	gin.SetMode(gin.ReleaseMode)
	h.g = gin.New()

	h.g.Use(gin.Recovery())

	h.g.Use(func(ctx *gin.Context) {
		h.Println(fmt.Sprintf("[%s] [%s]", ctx.Request.Method, ctx.Request.RequestURI))
		ctx.Next()
	})

	h.g.Use(otelgin.Middleware(ServiceName))

	h.g.GET("/reset", h.reset)
	h.g.GET("/metrics", h.metrics)
	h.g.GET("/details", h.DetailsHttp)

	h.g.Any("/functions/:entrypoint", h.ForwardToAgentHttp)
	h.g.Any("/functions/:entrypoint/*action", h.ForwardToAgentHttp)

	config := []transport.Config{transport.WithRpcPort(grpcPort), transport.WithDeployServer(h),
		transport.WithHttpPort(http), transport.WithHandler(h.g)}
	transport.Init(ctx, config...).Start()
}
