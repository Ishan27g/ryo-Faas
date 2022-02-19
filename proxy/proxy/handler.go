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
	"time"

	"github.com/Ishan27g/ryo-Faas/metrics"
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

type handler struct {
	g               *gin.Engine
	agent           map[string]transport.AgentWrapper // agentAddr : client
	functions       map[string]string                 // entrypoint : agentAddr
	httpFnProxyPort string

	ready chan string

	// metric *metrics.Functions
	*metrics.PrometheusMetrics

	proxies proxy
	*log.Logger

	jaegerPrv *metrics.JaegerProvider
}

func prettyJson(js interface{}) string {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(data)
}
func (h *handler) agentReady(address string) {
	go func() {
		h.ready <- address
	}()
}

func (h *handler) getFunctionAgent(entrypoint string) (transport.AgentWrapper, string) {
	if h.functions[entrypoint] == "" {
		h.Println("No agent to assigned to ", entrypoint)
		return nil, ""
	}
	return h.agent[h.functions[entrypoint]], h.functions[entrypoint]
}

func (h *handler) assignAgentToFunction(entrypoint string) (transport.AgentWrapper, string) {
	select {
	case agentAddr := <-h.ready:
		h.functions[entrypoint] = agentAddr
		return h.agent[h.functions[entrypoint]], agentAddr
	default:
		h.Println("No agent to assign to", entrypoint)
		return nil, ""
	}
}
func (h *handler) AddAgent(address string) bool {
	if address == "" {
		return false
	}
	c := transport.ProxyGrpcClient(address)
	if c == nil {
		return false
	}
	h.agent[address] = c
	h.agentReady(address)
	h.Println("Added new agent ", address)
	return true
}

func (h *handler) Deploy(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {

	now := time.Now()
	tr := otel.Tracer("Function deploy")
	ctxT, span := tr.Start(ctx, "proxy-deploy")
	span.SetAttributes(attribute.Key("entrypoint").String(request.Functions.Entrypoint))
	defer func() {
		span.SetAttributes(attribute.Key("add-proxy-rsp").String(time.Since(now).String()))
		span.End()
	}()

	h.Println("DEPLOY REQUEST FOR ", request.Functions)
	agent, address := h.assignAgentToFunction(request.Functions.Entrypoint)
	if agent == nil {
		return nil, errors.New("cannot assign an agent for" + request.Functions.Entrypoint)
	}
	defer h.agentReady(address)
	if !transport.UploadDir(agent, ctxT, request.Functions.Dir, request.Functions.Entrypoint) {
		return nil, errors.New("cannot upload directory to agent " + request.Functions.Entrypoint)
	}

	span.SetAttributes(attribute.Key("upload").String(time.Since(now).String()))
	now = time.Now()
	response := new(deploy.DeployResponse)
	agentRsp, err := agent.Deploy(ctxT, request)
	if err != nil {
		return nil, err
	}
	span.SetAttributes(attribute.Key("deploy").String(time.Since(now).String()))
	now = time.Now()

	for _, function := range agentRsp.Functions {
		jsonFn := types.RpcFunctionRspToJson(function)
		proxyUrl := h.proxies.add(jsonFn)
		function.Url = "http://" + h.httpFnProxyPort + proxyUrl
		response.Functions = append(response.Functions, function)

		span.SetAttributes(attribute.Key("entrypoint").String(function.Entrypoint))
		span.SetAttributes(attribute.Key("status").String(function.Status))
		span.SetAttributes(attribute.Key("url").String(function.Url))

	}
	h.Println("DEPLOY RESPONSE", prettyJson(response))
	return response, nil
}

func (h *handler) Stop(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {
	agent, address := h.getFunctionAgent(request.Functions.Entrypoint)
	if agent == nil {
		return nil, errors.New("cannot find agent for" + request.Functions.Entrypoint)
	}
	defer h.agentReady(address)
	response := new(deploy.DeployResponse)
	agentRsp, err := agent.Stop(ctx, request)
	if err != nil {
		return nil, err
	}
	for _, function := range agentRsp.Functions {
		h.proxies.remove(function.Entrypoint)
		function.Url = ""
		response.Functions = append(response.Functions, function)
	}
	return response, nil
}

func (h *handler) List(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	agent, address := h.getFunctionAgent(empty.GetEntrypoint())
	if agent == nil {
		return nil, errors.New("cannot find agent for" + empty.GetEntrypoint())
	}
	defer h.agentReady(address)
	response, err := agent.List(ctx, empty)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *handler) Details(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	var details *deploy.DeployResponse
	for _, client := range h.agent {
		detail, err := client.Details(ctx, empty)
		if err == nil {
			details.Functions = append(details.Functions, detail.Functions...)
		}
	}
	h.Println("Proxy details : ", h.proxies.details())
	h.Println("Function : ", h.functions)
	h.Println("Agents : ", h.agent)
	return details, nil
}

func (h *handler) Upload(server deploy.Deploy_UploadServer) error {

	return errors.New("no upload method at server. Call deploy")
}

func (h *handler) Logs(ctx context.Context, function *deploy.Function) (*deploy.Logs, error) {
	agent, address := h.getFunctionAgent(function.Entrypoint)
	if agent == nil {
		return nil, errors.New("cannot find agent for" + function.Entrypoint)
	}
	defer h.agentReady(address)
	response, err := agent.Logs(ctx, function)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *handler) ForwardToAgentHttp(c *gin.Context) {
	fnName := c.Param("entrypoint")
	//done := h.metric.Invoked(fnName)

	//defer close(done)
	var statusCode = http.StatusBadGateway
	var agent = "Not found"

	// h.jaegerPrv.Tracer(MetricTracerFwToAgent).Start(c.Request.Context(), fnName)
	sp := trace.SpanFromContext(c.Request.Context())
	sp.SetAttributes(attribute.Key("forward-to").String(fnName))

	now := time.Now()
	if proxy, fnServiceHost, atAgent := h.proxies.get(fnName); fnServiceHost != "" {
		statusCode, span := proxy.ServeHTTP(c.Writer, c.Request, fnServiceHost)
		agent = atAgent
		span.SetAttributes(attribute.Key("status").String(strconv.Itoa(statusCode)))
		span.SetAttributes(attribute.Key("round-trip").String(time.Since(now).String()))
	} else {
		c.String(http.StatusBadGateway, "Not found - ", fnName)
	}
	// done <- statusCode >= http.StatusOK && statusCode <= http.StatusAccepted
	h.PrometheusMetrics.Update(fnName, agent, statusCode)
}

func (h *handler) AgentJoinHttp(c *gin.Context) {
	agentAddr := c.Query("address")
	if agentAddr != "" && h.AddAgent(agentAddr) {
		c.String(200, "Joined")
		return
	}
	c.String(400, "Cannot connect to address -"+agentAddr+"\n")
}

func (h *handler) DetailsHttp(c *gin.Context) {
	rsp := make(map[string]interface{})
	rsp["Proxy details"] = h.proxies.details()
	rsp["Function"] = h.functions
	rsp["Agents"] = h.agent
	c.JSON(200, rsp)
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

func (h *handler) ListHttp(c *gin.Context) {
	var rsp []types.FunctionJsonRsp
	for entryPoint := range h.functions {
		list, err := h.List(c, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: entryPoint}})
		if err == nil {
			for _, function := range list.Functions {
				rsp = append(rsp, types.RpcFunctionRspToJson(function))
			}
		}
	}
	c.JSON(200, rsp)
}

func (h *handler) MetricsDevHttp(c *gin.Context) {
	var agentMetrics []map[string]*metrics.Functions
	proxyMetrics := make(map[string]*metrics.Metric)
	for _, wrapper := range h.agent {
		agentMetrics = append(agentMetrics, wrapper.GetMetrics())
	}
	//for entryPoint, metric := range h.metric.Fns {
	//	proxyMetrics[entryPoint] = metric
	//	fn := h.proxies.functions[entryPoint]
	//	proxyMetrics[entryPoint].Function.AtAgent = fn.agentAddr
	//	proxyMetrics[entryPoint].Function.Url = "http://" + h.httpFnProxyPort + fn.proxyFrom
	//	proxyMetrics[entryPoint].Function.ProxyServiceAddr = fn.proxyTo
	//}
	c.JSON(200, &gin.H{"agentMetrics": agentMetrics, "proxyMetrics": proxyMetrics})
}

func (h *handler) MetricsPrometheus(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func Start(ctx context.Context, grpcPort, http string, provider *metrics.JaegerProvider) {
	h := new(handler)
	h.agent = make(map[string]transport.AgentWrapper)
	h.functions = make(map[string]string)
	h.httpFnProxyPort = http
	h.proxies = newProxy()
	h.Logger = log.New(os.Stdout, "[PROXY-HANDLER] ", log.Ltime)
	h.ready = make(chan string)

	//h.metric = metrics.NewMetricMap()
	h.PrometheusMetrics = metrics.InitPrometheus()

	h.jaegerPrv = provider

	gin.SetMode(gin.DebugMode)
	h.g = gin.New()
	h.g.Use(gin.Recovery())

	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gin-gonic/gin/otelgin/example/server.go
	h.g.Use(otelgin.Middleware("proxy-server"))

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
	// curl  http://localhost:9002/addAgent?address=1
	h.g.GET("/addAgent", h.AgentJoinHttp)
	// curl  http://localhost:9002/list
	h.g.GET("/list", h.ListHttp)
	// curl  http://localhost:9002/metrics
	h.g.GET("/metrics", h.MetricsPrometheus)
	h.g.GET("/metricsDev", h.MetricsDevHttp)

	// curl -X POST http://localhost:9002/functions/method1 -H 'Content-Type: application/json' -d '{
	//    "data": "http://host.docker.internal:9002/functions/method2"
	//  }'
	h.g.Any("/functions/:entrypoint", h.ForwardToAgentHttp) // todo method=ANY

	//h.g.GET("/stop/:entrypoint", h.StopHttp)
	//h.g.POST("/file/:entrypoint", h.UploadHttp)
	//h.g.GET("/log", h.LogHttp) // /log?entrypoint=

	transport.Init(ctx, h, grpcPort, h.g, http).Start()
}
