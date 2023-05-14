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

	"github.com/Ishan27g/noware/pkg/noop"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/pkg/docker"
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/scale"
	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName = "rfa-proxy"

	DefaultRpc           = ":9998"
	DefaultHttp          = ":9999"
	DeployedFunctionPort = ":6000"

	LocalProxy = "LOCAL_PROXY"
	LocalDB    = "LOCAL_DB"
)

var isProxyLocal = func() bool { return os.Getenv(LocalProxy) == "YES" }
var isDbLocal = func() bool { return os.Getenv(LocalDB) == "YES" }

var scaleEndpoint = hn() + DefaultHttp + Functions + "/scale"

var hn = func() string {
	hostname, _ := os.Hostname()
	if strings.Contains(hostname, "Ishan") { // if running locally
		return "http://" + "localhost"
	}
	// if in docker network
	return "http://host.docker.internal"
}
var buildHostName = func(entrypoint string) string {
	hostname, _ := os.Hostname()
	if strings.Contains(hostname, "Ishan") { // if running locally
		return "http://" + "localhost"
	}
	// if in docker network
	return "http://" + "rfa-deploy-" + strings.ToLower(entrypoint)
}

type handler struct {
	g               *gin.Engine
	httpFnProxyPort string
	proxies         proxy
	*scale.Monitor
	*log.Logger
}

func prettyJson(js interface{}) string {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(data)
}
func addFunctionAttributes(span *trace.Span, function *deploy.Function) {
	(*span).SetAttributes(attribute.Key(tracing.Entrypoint).String(function.Entrypoint))
	(*span).SetAttributes(attribute.Key(tracing.Url).String(function.Url))
	(*span).SetAttributes(attribute.Key(tracing.Status).String(function.Status))
	(*span).SetAttributes(attribute.Key(tracing.IsMain).Bool(function.IsMain))
	(*span).SetAttributes(attribute.Key(tracing.IsAsync).Bool(function.Async))
}

func updateSpan(sp trace.Span, deploymentType string, statusCode int, now time.Time, fnName string) trace.Span {
	sp.SetAttributes(attribute.Key("deployment").String(deploymentType))
	sp.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
	sp.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
	sp.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(statusCode))))
	return sp
}
func (h *handler) Deploy(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {

	span := trace.SpanFromContext(ctx)
	instance := 0
	fnName := strings.ToLower(request.Functions[0].Entrypoint)

	var found, isAsync, isMain bool
	if found, isAsync, isMain = h.proxies.getFlags(fnName); found {
		// scale-up this function container by 1
		instance = len(h.proxies.groups[fnName].urls)
		if instance == 0 {
			instance = 1
		}
		// don't consider cli flags for scaling
		// deploy with flags are as they were before,
		for _, function := range request.Functions {
			function.IsMain = isMain
			function.Async = isAsync
		}
	}
	d := docker.New()
	if isProxyLocal() {
		d.SetLocalProxy()
	}
	if isDbLocal() {
		d.SetLocalDb()
	}
	err := d.RunFunctionInstance(request.Functions[0].Entrypoint, instance)
	if err != nil {
		span.AddEvent("unable to start container for " + request.Functions[0].Entrypoint)
		h.Println(err.Error())
		return nil, errors.New("unable to start container for " + request.Functions[0].Entrypoint)
	}
	span.SetAttributes(attribute.Key("entrypoint").String(fnName))
	span.SetAttributes(attribute.Key("instance").Int(instance))

	response := new(deploy.DeployResponse)

	hnFn := buildHostName(request.Functions[0].Entrypoint+strconv.Itoa(instance)) + DeployedFunctionPort
	for _, function := range request.Functions {
		function.ProxyServiceAddr = hnFn
		jsonFn := types.RpcFunctionRspToJson(function)

		proxyUrl := h.proxies.add(jsonFn, instance)
		function.Url = "http://localhost" + h.httpFnProxyPort + proxyUrl
		response.Functions = append(response.Functions, function)

		// h.UselessMetrics.Register(function)
		addFunctionAttributes(&span, function)
	}

	// 	h.Println("DEPLOY RESPONSE IS", prettyJson(response))
	span.AddEvent(prettyJson(response))
	return response, nil
}

func (h *handler) Stop(ctx context.Context, request *deploy.Empty) (*deploy.DeployResponse, error) {
	response := new(deploy.DeployResponse)

	span := trace.SpanFromContext(ctx)

	fnName := strings.ToLower(request.GetEntrypoint())

	instance := h.proxies.remove(fnName)

	span.SetAttributes(attribute.Key("entrypoint").String(fnName))
	span.SetAttributes(attribute.Key("instance").Int(instance))

	if docker.New().StopFunctionInstance(fnName, instance) != nil {
		h.Println("Unable to stop container", fnName, " for instance ", instance)
		span.AddEvent("Unable to stop container" + fnName + " for instance " + strconv.Itoa(instance))
		return response, nil
	}
	span.AddEvent("Stopped container " + fnName + " for instance " + strconv.Itoa(instance))
	h.Println("Stopped container " + fnName + " for instance " + strconv.Itoa(instance))

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

func (h *handler) ForwardToAgentHttp(c *gin.Context) {
	fnName := c.Param("entrypoint")
	var statusCode = http.StatusBadGateway
	var ctxR context.Context

	sp, ctx := trace.SpanFromContext(c.Request.Context()), c.Request.Context()

	if !sp.IsRecording() {
		ctxR, sp = otel.Tracer(ServiceName).Start(ctx, "forward"+"-"+fnName)
	} else {
		ctxR = trace.ContextWithSpan(ctx, sp)
	}
	sp.SetName(fnName)
	sp.SetAttributes(attribute.KeyValue{
		Key:   "noop",
		Value: attribute.BoolValue(noop.ContainsNoop(ctxR)),
	})

	newReq := newFwRequestWithCtx(ctxR, c.Request)
	now := time.Now()

	if proxy, fnServiceHost, isAsync, isMain := h.proxies.invoke(fnName); fnServiceHost != "" {

		h.Monitor.Invoked(fnName)

		sp.SetAttributes(semconv.HTTPHostKey.String(fnServiceHost))
		if isAsync {
			if noop.ContainsNoop(ctxR) {
				return
			}
			FuncFw.NewAsyncNats(fnName, "").HandleAsyncNats(c.Writer, newReq)
			statusCode = http.StatusOK
			sp = updateSpan(sp, "async-nats", statusCode, now, fnName)
		} else {
			var stc int
			var span trace.Span
			if isMain {
				if noop.ContainsNoop(ctxR) {
					return
				}
				stc, span = proxy.ServeHTTP(c.Writer, newReq, fnServiceHost, strings.ToLower(fnName))
			} else {
				if noop.ContainsNoop(ctxR) {
					return
				}
				stc, span = proxy.ServeHTTP(c.Writer, newReq, fnServiceHost, "")
			}
			statusCode = stc
			span = updateSpan(span, "http", statusCode, now, fnName)
		}
	} else {
		sp.SetAttributes(attribute.Key("No Proxy found").String(fnName))
		c.String(http.StatusBadGateway, "Not found - "+fnName)
	}
}

func (h *handler) reset(c *gin.Context) {
	h.proxies = newProxy()
	c.Status(http.StatusAccepted)
	h.Println("reset")
}

func (h *handler) DetailsHttp(c *gin.Context) {

	span := trace.SpanFromContext(c)

	var details []types.FunctionJsonRsp

	for _, fn := range h.proxies.details() {
		upstream := h.proxies.getUpstreamFor(fn.Name)
		if checkHealth(upstream) {
			pFn := h.proxies.asDefinition(fn.Name)
			details = append(details, types.FunctionJsonRsp{
				Name:      pFn.fnName,
				Proxy:     upstream,
				IsAsync:   pFn.isAsync,
				IsMain:    pFn.isMain,
				Instances: len(h.proxies.groups[fn.Name].urls),
			})

			span.SetAttributes(attribute.Key(fn.Name).String(prettyJson(pFn)))

		} else {
			h.Println(upstream + " unreachable")
			span.SetAttributes(attribute.Key(fn.Name).String("unreachable"))
		}
	}
	span.AddEvent(prettyJson(details))

	c.JSON(200, details)
}

func (h *handler) SwitchMetrics(c *gin.Context) {
	h.Println("SwitchMetrics -> query param -> ", c.Query("bool"))
	if c.Query("bool") == "" {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	startExporter := strings.EqualFold(c.Query("bool"), "true")
	if startExporter {
		scale.StartExporter(h.Monitor, scaleEndpoint)
	} else {
		scale.StopExporter()
	}
	c.JSON(http.StatusOK, startExporter)
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
	h.Logger = log.New(os.Stdout, ServiceName, log.Ltime)
	h.httpFnProxyPort = http
	h.proxies = newProxy()
	h.Monitor = scale.NewMetricsMonitor()

	scale.StartExporter(h.Monitor, scaleEndpoint)

	FuncFw.Export.HttpGin("Reset", "/reset", h.reset)
	FuncFw.Export.HttpGin("DetailsHttp", "/details", h.DetailsHttp)
	FuncFw.Export.HttpGin("SwitchMetrics", "/metrics/:bool", h.SwitchMetrics)
	FuncFw.Export.HttpGin("ForwardToAgentHttp", "/functions/:entrypoint", h.ForwardToAgentHttp)

	config := []transport.Config{transport.WithRpcPort(grpcPort), transport.WithDeployServer(h)}
	transport.Init(ctx, config...).Start()
	FuncFw.Start(strings.TrimPrefix(http, ":"), ServiceName)
}
