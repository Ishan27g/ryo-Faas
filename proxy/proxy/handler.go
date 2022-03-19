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
	agent           map[string]transport.AgentWrapper // agentAddr : client
	functions       map[string]string                 // entrypoint : agentAddr
	httpFnProxyPort string

	ready chan string

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

func (h *handler) assignAgentToFunction(fns []*deploy.Function) (transport.AgentWrapper, string) {
	select {
	case agentAddr := <-h.ready:
		for _, fn := range fns {
			h.functions[fn.Entrypoint] = agentAddr
		}
		return h.agent[h.functions[fns[0].Entrypoint]], agentAddr
	default:
		h.Println("No agent to assign")
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
	span := trace.SpanFromContext(ctx)

	agent, address := h.assignAgentToFunction(request.Functions)
	if agent == nil {
		return nil, errors.New("cannot assign an agent")
	}
	defer h.agentReady(address)
	for _, function := range request.Functions {
		span.SetAttributes(attribute.Key("entrypoint").String(function.Entrypoint))
		ctx = trace.ContextWithSpan(ctx, span)
		if !transport.UploadDir(agent, ctx, function.Dir, function.Entrypoint) {
			span.AddEvent("Upload error", trace.WithAttributes(attribute.Key(function.Dir).String(function.AtAgent)))
			return nil, errors.New("cannot upload directory to agent " + function.Entrypoint)
		}

		span = trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.Key("upload-duration").String(time.Since(now).String()))

		now = time.Now()
	}

	ctx = trace.ContextWithSpan(ctx, span)
	response := new(deploy.DeployResponse)
	agentRsp, err := agent.Deploy(ctx, request)
	if err != nil {
		span.AddEvent("Deploy error at agent")
		return nil, err
	}
	span = trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.Key("deploy-duration").String(time.Since(now).String()))
	now = time.Now()

	for _, function := range agentRsp.Functions {
		jsonFn := types.RpcFunctionRspToJson(function)
		proxyUrl := h.proxies.add(jsonFn)
		function.Url = "http://" + h.httpFnProxyPort + proxyUrl
		response.Functions = append(response.Functions, function)
	}
	span.AddEvent(prettyJson(response))
	h.Println("DEPLOY RESPONSE", prettyJson(response))
	return response, nil
}

func (h *handler) Stop(ctx context.Context, request *deploy.Empty) (*deploy.DeployResponse, error) {

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Key("entrypoint").String(request.GetEntrypoint()))

	entryPoint := request.GetEntrypoint()
	agent, address := h.getFunctionAgent(entryPoint)
	if agent == nil {
		span.AddEvent("cannot find agent for" + entryPoint)
		return nil, errors.New("cannot find agent for" + entryPoint)
	}
	defer h.agentReady(address)

	ctx = trace.ContextWithSpan(ctx, span)
	agentRsp, err := agent.Stop(ctx, request)
	if err != nil {
		span.AddEvent("Stop error at agent", trace.WithAttributes(attribute.Key(address).String(request.GetEntrypoint())))
		return nil, err
	}

	response := new(deploy.DeployResponse)
	for _, function := range agentRsp.Functions {
		h.proxies.remove(function.Entrypoint)
		function.Url = ""
		response.Functions = append(response.Functions, function)
	}
	return response, nil
}

func (h *handler) List(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Key("entrypoint").String(empty.GetEntrypoint()))

	agent, address := h.getFunctionAgent(empty.GetEntrypoint())
	if agent == nil {
		span.AddEvent("cannot find agent for" + empty.GetEntrypoint())
		return nil, errors.New("cannot find agent for" + empty.GetEntrypoint())
	}
	defer h.agentReady(address)

	ctx = trace.ContextWithSpan(ctx, span)

	response, err := agent.List(ctx, empty)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)

	span = trace.SpanFromContext(ctx)
	for _, fn := range response.Functions {
		fn.Url = h.proxies.functions[strings.ToLower(fn.Entrypoint)].proxyFrom

	}
	span.AddEvent(prettyJson(response.Functions))
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
	span := trace.SpanFromContext(ctx)
	span.AddEvent(prettyJson(h.proxies.details()))
	span.AddEvent(prettyJson(h.functions))
	span.AddEvent(prettyJson(h.agent))

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

	var statusCode = http.StatusBadGateway

	// get from request
	sp := trace.SpanFromContext(c.Request.Context())
	sp.SetAttributes(attribute.Key("entrypoint").String(fnName))
	ctxR := trace.ContextWithSpan(c.Request.Context(), sp)

	now := time.Now()
	if proxy, fnServiceHost, atAgent, _ := h.proxies.get(fnName); fnServiceHost != "" {
		//if isAsyncNats {
		//	FuncFw.NewAsyncNats(fnName, "").HandleAsyncNats(c.Writer, c.Request)
		//	sp.SetAttributes(attribute.Key("function-ASYNC-NATS").String(atAgent))
		//	sp.SetAttributes(attribute.Key("function-at-agent").String(atAgent))
		//	sp.SetAttributes(attribute.Key("function-rsp-status").String(strconv.Itoa(statusCode)))
		//	sp.SetAttributes(attribute.Key("function-round-trip").String(time.Since(now).String()))
		//	sp.AddEvent(fnName, trace.WithAttributes(attribute.Key(fnName).String(strconv.Itoa(statusCode))))
		//} else {
		stc, span := proxy.ServeHTTP(ctxR, c.Writer, c.Request, fnServiceHost)
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

func (h *handler) AgentJoinHttp(c *gin.Context) {
	agentAddr := c.Query("address")
	if agentAddr != "" && h.AddAgent(agentAddr) {
		c.String(200, "Joined")
		return
	}
	c.String(400, "Cannot connect to address -"+agentAddr+"\n")
}

func (h *handler) DetailsHttp(c *gin.Context) {

	var details []types.FunctionJsonRsp
	fmt.Println(h.proxies.functions)
	tr := otel.Tracer("Function list")
	ctxT, span := tr.Start(c.Request.Context(), "proxy-details")

	defer func() {
		span.End()
	}()

	for _, fn := range h.proxies.details() {
		fmt.Println(fn)
		entrypoint := strings.ToLower(fn.Name)
		pFn := h.proxies.getFn(entrypoint)
		if pFn == nil {
			continue
		}
		client, _ := h.getFunctionAgent(fn.Name)
		if client == nil {
			continue
		}

		detail, err := client.Details(ctxT, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: fn.Name}})
		if err == nil {
			for _, fn := range detail.Functions {
				fn.Url = "http://" + h.httpFnProxyPort + pFn.proxyFrom
				details = append(details, types.RpcFunctionRspToJson(fn))
				span.SetAttributes(attribute.Key(fn.Entrypoint).String(fn.Status))
			}
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

func Start(ctx context.Context, grpcPort, http string, agents ...string) {
	h := new(handler)
	h.agent = make(map[string]transport.AgentWrapper)
	h.functions = make(map[string]string)
	h.httpFnProxyPort = http
	h.proxies = newProxy()
	h.Logger = log.New(os.Stdout, "[PROXY-HANDLER] ", log.Ltime)
	h.ready = make(chan string)

	for _, v := range agents {
		if !h.AddAgent(v) {
			log.Fatal("Unable to add agent ", v)
		}
	}

	gin.SetMode(gin.ReleaseMode)
	h.g = gin.New()
	h.g.Use(gin.Recovery())

	h.g.Use(otelgin.Middleware("proxy-server"))

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
	// curl  http://localhost:9002/addAgent?address=1
	h.g.GET("/addAgent", h.AgentJoinHttp)
	// curl  http://localhost:9002/list
	h.g.GET("/list", h.ListHttp)
	// curl  http://localhost:9002/metrics
	h.g.GET("/metrics", h.MetricsPrometheus)

	// curl -X POST http://localhost:9002/functions/method1 -H 'Content-Type: application/json' -d '{
	//    "data": "http://host.docker.internal:9002/functions/method2"
	//  }'
	h.g.Any("/functions/:entrypoint", h.ForwardToAgentHttp)

	//h.g.GET("/stop/:entrypoint", h.StopHttp)
	//h.g.POST("/file/:entrypoint", h.UploadHttp)
	//h.g.GET("/log", h.LogHttp) // /log?entrypoint=

	transport.Init(ctx, struct {
		IsDeploy bool
		Server   interface{}
	}{IsDeploy: true, Server: h}, grpcPort, h.g, http).Start()

}
