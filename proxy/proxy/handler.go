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

	"github.com/Ishan27g/ryo-Faas/docker"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const MetricTracerFwToAgent = "proxy-function-call"
const UrlLookup = "proxy-lookup"
const HttpProxy = "http-proxy"
const TracerName = "proxy"

const DefaultRpc = ":9998"
const DefaultHttp = ":9999"

type handler struct {
	g               *gin.Engine
	httpFnProxyPort string

	proxies proxy
	*log.Logger
}

func getTracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(TracerName)
}
func prettyJson(js interface{}) string {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(data)
}
func (h *handler) Deploy(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {

	response := new(deploy.DeployResponse)

	for _, function := range request.Functions {
		hnFn := "http://" + "rfa-deploy-" + strings.ToLower(function.Entrypoint) + ":6000"
		function.ProxyServiceAddr = hnFn // + function.Entrypoint
		jsonFn := types.RpcFunctionRspToJson(function)
		proxyUrl := h.proxies.add(jsonFn)
		function.Url = "http://localhost" + h.httpFnProxyPort + proxyUrl
		response.Functions = append(response.Functions, function)
	}
	h.Println("DEPLOY RESPONSE IS", prettyJson(response))
	return response, nil
}

func (h *handler) Stop(ctx context.Context, request *deploy.Empty) (*deploy.DeployResponse, error) {

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Key("entrypoint").String(request.GetEntrypoint()))
	h.proxies.remove(request.GetEntrypoint())

	if docker.New().StopFunction(request.GetEntrypoint()) != nil {
		h.Println("Unable to stop ", request.GetEntrypoint())
	}

	response := new(deploy.DeployResponse)
	return response, nil
}

func (h *handler) List(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	response := new(deploy.DeployResponse)
	return response, nil
}
func (h *handler) Details(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	var details *deploy.DeployResponse

	span := trace.SpanFromContext(ctx)
	span.AddEvent(prettyJson(h.proxies.details()))

	h.Println("Proxy details : ", h.proxies.details())
	return details, nil
}

func (h *handler) Upload(stream deploy.Deploy_UploadServer) error {

	return errors.New("no upload method")

}

func (h *handler) Logs(ctx context.Context, function *deploy.Function) (*deploy.Logs, error) {
	r := new(deploy.Logs)
	return r, nil
}

func (h *handler) ForwardToAgentHttp(c *gin.Context) {
	fnName := c.Param("entrypoint")

	var statusCode = http.StatusBadGateway

	// get from request
	sp := trace.SpanFromContext(c.Request.Context())
	sp.SetAttributes(attribute.Key("entrypoint").String(fnName))
	ctxR := trace.ContextWithSpan(c.Request.Context(), sp)

	now := time.Now()
	if proxy, fnServiceHost, atAgent, _, isMain := h.proxies.get(fnName); fnServiceHost != "" {
		//if isAsyncNats {
		//	FuncFw.NewAsyncNats(fnName, "").HandleAsyncNats(c.Writer, c.Request)
		//	sp.SetAttributes(attribute.Key("function-ASYNC-NATS").String(atAgent))
		//	sp.SetAttributes(attribute.Key("function-at-agent").String(atAgent))
		//	sp.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
		//	sp.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
		//	sp.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(statusCode))))
		//} else {
		var stc int
		var span trace.Span
		if isMain {
			stc, span = proxy.ServeHTTP(ctxR, c.Writer, c.Request, fnServiceHost, strings.ToLower(fnName))
		} else {
			stc, span = proxy.ServeHTTP(ctxR, c.Writer, c.Request, fnServiceHost, "")
		}
		statusCode = stc
		span.SetAttributes(attribute.Key("function-at-agent").String(atAgent))
		span.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
		span.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
		span.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(statusCode))))
		//}
	} else {
		c.String(http.StatusBadGateway, "Not found - "+fnName)
	}
}

func (h *handler) reset(c *gin.Context) {
	h.proxies = newProxy()
	c.Status(http.StatusAccepted)
	h.Println("reset")
}
func (h *handler) DetailsHttp(c *gin.Context) {

	var details []types.FunctionJsonRsp

	for _, fn := range h.proxies.details() {
		if checkHealth(h.proxies.getFuncFwHost(fn.Name)) {
			pFn := h.proxies.asDefinition(fn.Name)
			details = append(details, types.FunctionJsonRsp{
				Name:    pFn.fnName,
				Proxy:   pFn.proxyTo,
				IsAsync: pFn.isAsyncNats,
			})
		} else {
			h.Println(h.proxies.getFuncFwHost(fn.Name) + " unreachable")
		}
	}
	fmt.Println(details)
	c.JSON(200, details)
}

func (h *handler) DeployHttp(c *gin.Context) {
	var req []types.FunctionJson
	var r []types.FunctionJsonRsp
	err := c.ShouldBindJSON(&req)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(400, nil)
		return
	}
	ctx, can := context.WithTimeout(c, 2*time.Second)
	defer can()
	for _, json := range req {
		response, err := h.Deploy(ctx, &deploy.DeployRequest{Functions: types.JsonFunctionToRpc(json)})
		if err != nil {
			return
		}
		for _, function := range response.Functions {
			r = append(r, types.RpcFunctionRspToJson(function))
		}
	}
	c.JSON(200, r)
}

func (h *handler) MetricsPrometheus(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func (h *handler) DeployHttpAsync(c *gin.Context) {
	var req []types.FunctionJson
	var r []types.FunctionJsonRsp
	err := c.ShouldBindJSON(&req)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(400, nil)
		return
	}
	ctx, can := context.WithTimeout(c, 2*time.Second)
	defer can()
	for _, json := range req {
		fn := types.JsonFunctionToRpc(json)
		fn[0].Async = true
		response, err := h.Deploy(ctx, &deploy.DeployRequest{Functions: fn})
		if err != nil {
			return
		}
		for _, function := range response.Functions {
			r = append(r, types.RpcFunctionRspToJson(function))
		}
	}
	c.JSON(200, r)
}
func checkHealth(addr string) bool {
	resp, err := http.Get(addr + "/healthcheck")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func Start(ctx context.Context, grpcPort, http string, agents ...string) {
	h := new(handler)
	h.httpFnProxyPort = http
	h.proxies = newProxy()
	h.Logger = log.New(os.Stdout, "[PROXY-HANDLER] ", log.Ltime)

	gin.SetMode(gin.ReleaseMode)
	h.g = gin.New()
	h.g.Use(gin.Recovery())

	h.g.Use(func(ctx *gin.Context) {
		h.Println(fmt.Sprintf("[%s] [%s]", ctx.Request.Method, ctx.Request.RequestURI))
		ctx.Next()
	})

	h.g.Use(otelgin.Middleware("proxy-server"))

	h.g.GET("/reset", h.reset)

	// proxy curl -X POST http://localhost:9002/deploy -H 'Content-Type: application/json' -d '[
	//  {
	//    "packageDir": "/Users/ishan/Desktop/multi/method1",
	//    "name" : "Method1",
	//    "filePath": "/Users/ishan/Desktop/multi/method1.go"
	//  }
	//]'
	h.g.POST("/deploy/async", h.DeployHttpAsync)

	h.g.POST("/deploy", h.DeployHttp)
	// curl  http://localhost:9002/details
	h.g.GET("/details", h.DetailsHttp)
	// curl  http://localhost:9002/list
	// curl  http://localhost:9002/metrics
	h.g.GET("/metrics", h.MetricsPrometheus)

	// curl -X POST http://localhost:9002/functions/method1 -H 'Content-Type: application/json' -d '{
	//    "data": "http://host.docker.internal:9002/functions/method2"
	//  }'
	h.g.Any("/functions/:entrypoint", h.ForwardToAgentHttp)
	h.g.Any("/functions/:entrypoint/*action", h.ForwardToAgentHttp)

	//h.g.GET("/stop/:entrypoint", h.StopHttp)
	//h.g.POST("/file/:entrypoint", h.UploadHttp)
	//h.g.GET("/log", h.LogHttp) // /log?entrypoint=

	transport.Init(ctx, struct {
		IsDeploy bool
		Server   interface{}
	}{IsDeploy: true, Server: h}, grpcPort, h.g, http).Start()

}
