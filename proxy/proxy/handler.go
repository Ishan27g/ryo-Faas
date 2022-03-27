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
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const TracerFwToAgent = "proxy-function-call"
const ServiceName = "rfa-proxy"

const DefaultRpc = ":9998"
const DefaultHttp = ":9999"

type handler struct {
	g               *gin.Engine
	httpFnProxyPort string

	proxies proxy
	*log.Logger
}

func getTracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(ServiceName)
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
	for _, function := range request.Functions {
		span.SetAttributes(attribute.Key("entrypoint").String(function.GetEntrypoint()))
	}

	response := new(deploy.DeployResponse)
	var buildHostName = func(entrypoint string) string {
		hostname, _ := os.Hostname()
		if strings.Contains(hostname, "Ishan") {
			// if running locally
			return "localhost"
		}
		// if in docker network
		return "rfa-deploy-" + strings.ToLower(request.Functions[0].Entrypoint)
	}
	hnFn := "http://" + buildHostName(request.Functions[0].Entrypoint) + ":6000"
	for _, function := range request.Functions {
		function.ProxyServiceAddr = hnFn // + function.Entrypoint
		jsonFn := types.RpcFunctionRspToJson(function)
		proxyUrl := h.proxies.add(jsonFn)
		function.Url = "http://localhost" + h.httpFnProxyPort + proxyUrl
		response.Functions = append(response.Functions, function)

		span.SetAttributes(attribute.Key(function.Entrypoint).String(prettyJson(function)))

	}
	h.Println("DEPLOY RESPONSE IS", prettyJson(response))
	return response, nil
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

func (h *handler) Details(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {

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
		span.SetAttributes(attribute.Key(df.Entrypoint).String(prettyJson(*df)))
		details.Functions = append(details.Functions)
	}
	h.Println("Proxy details : ", h.proxies.details())
	return details, nil
}

func (h *handler) Upload(stream deploy.Deploy_UploadServer) error {
	return errors.New("no upload method")
}

func (h *handler) ForwardToAgentHttp(c *gin.Context) {
	fnName := c.Param("entrypoint")

	var statusCode = http.StatusBadGateway

	// get from request
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
	if proxy, fnServiceHost, atAgent, isAsyncNats, isMain := h.proxies.get(fnName); fnServiceHost != "" {
		if isAsyncNats {
			FuncFw.NewAsyncNats(fnName, "").HandleAsyncNats(c.Writer, newReq)
			sp.SetAttributes(attribute.Key("function-async-nats").String(fnName))
			sp.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
			sp.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
			sp.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(http.StatusAccepted))))
		} else {
			var stc int
			var span trace.Span
			if isMain {
				stc, span = proxy.ServeHTTP(ctxR, c.Writer, newReq, fnServiceHost, strings.ToLower(fnName))
			} else {
				stc, span = proxy.ServeHTTP(ctxR, c.Writer, newReq, fnServiceHost, "")
			}
			statusCode = stc
			span.SetAttributes(attribute.Key("function-http").String(atAgent))
			span.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
			span.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
			span.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(statusCode))))
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
		if checkHealth(h.proxies.getFuncFwHost(fn.Name)) {
			pFn := h.proxies.asDefinition(fn.Name)
			details = append(details, types.FunctionJsonRsp{
				Name:    pFn.fnName,
				Proxy:   pFn.proxyTo,
				IsAsync: pFn.isAsyncNats,
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

	// proxy curl -X POST http://localhost:9002/deploy -H 'Content-Type: application/json' -d '[
	//  {
	//    "packageDir": "/Users/ishan/Desktop/multi/method1",
	//    "name" : "Method1",
	//    "filePath": "/Users/ishan/Desktop/multi/method1.go"
	//  }
	//]'
	h.g.POST("/deploy", h.DeployHttp)
	// curl  http://localhost:9002/details
	h.g.GET("/details", h.DetailsHttp)
	// curl  http://localhost:9002/list

	// curl -X POST http://localhost:9002/functions/method1 -H 'Content-Type: application/json' -d '{
	//    "data": "http://host.docker.internal:9002/functions/method2"
	//  }'
	h.g.Any("/functions/:entrypoint", h.ForwardToAgentHttp)
	h.g.Any("/functions/:entrypoint/*action", h.ForwardToAgentHttp)

	transport.Init(ctx, struct {
		IsDeploy bool
		Server   interface{}
	}{IsDeploy: true, Server: h}, grpcPort, h.g, http).Start()

}
