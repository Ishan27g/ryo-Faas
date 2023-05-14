package FuncFw

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/Ishan27g/noware/pkg/middleware"
	database "github.com/Ishan27g/ryo-Faas/database/client"
	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var (
	serviceName                  = ""
	jaegerHost                   = os.Getenv("JAEGER")
	zipkinHost                   = os.Getenv("ZIPKIN")
	databaseAddress              = os.Getenv("DATABASE")
	httpSrv         *http.Server = nil
	logger                       = log.New(os.Stdout, "[func-fw]", log.LstdFlags)
	healthCheckUrl               = "/healthcheck"
	stopUrl                      = "/stop"
	Export                       = funcFw{
		httpFns:       make(map[string]*httpFunction),
		httpFnsGin:    make(map[string]*httpFnGin),
		httpAsync:     make(map[string]*HttpAsync),
		httpAsyncNats: make(map[string]*HttpAsync),
		storeEvents:   nil,
		funcCtx:       newCtx[any](nil),
	}
	provider tracing.TraceProvider
)

func Start(port, service string) {
	if service == "" {
		serviceName, _ = os.Hostname()
	} else {
		serviceName = service
	}

	// health check
	Export.Http("Healthcheck", healthCheckUrl, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// stop
	Export.Http("StopFuncFw", stopUrl, func(w http.ResponseWriter, r *http.Request) {
		defer Stop()
		w.WriteHeader(http.StatusOK)
	})

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(otelgin.Middleware(serviceName))
	g.Use(middleware.Gin())

	g.Use(func(ctx *gin.Context) {
		fmt.Println(fmt.Sprintf("\t\t\t[%s] [%s]", ctx.Request.Method, ctx.Request.RequestURI))
		ctx.Next()
	})

	if jaegerHost == "" && zipkinHost != "" {
		provider = tracing.Init("zipkin", serviceName, serviceName)
	}
	if zipkinHost == "" && jaegerHost != "" {
		provider = tracing.Init("jaeger", serviceName, serviceName)
	}
	if provider == nil {
		provider = tracing.Init("jaeger", serviceName, serviceName)
	}

	// apply store event handlers
	if Export.storeEvents != nil {
		if database.Connect(databaseAddress) == nil {
			log.Fatal("Store: Cannot connect to database")
		}
		if !applyEvents() {
			log.Fatal("Store: Unable to apply event cbs")
		}
	}

	// apply http handlers
	for _, function := range Export.getHttp() {
		g.Any(function.UrlPath, function.AsGin())
		g.Any(function.UrlPath+"/*any", function.AsGin())
		logger.Println("[http] " + function.Entrypoint + " at " + function.UrlPath)
	}

	for _, httpAsync := range Export.getHttpAsync() {
		g.Any(httpAsync.UrlPath, httpAsync.AsGin())
		g.Any(httpAsync.UrlPath+"/*any", httpAsync.AsGin())
		logger.Println("[http-Async] " + httpAsync.Entrypoint + " at " + httpAsync.UrlPath)
	}

	// apply http gin handlers
	for _, function := range Export.getHttpGin() {
		g.Any(function.UrlPath, function.gf)
		g.Any(function.UrlPath+"/*any", function.gf)
		logger.Println("[http-gin] " + function.Entrypoint + " at " + function.UrlPath)
	}

	// apply http async nats handlers
	for _, httpAsync := range Export.getHttpAsyncNats() {
		an := NewAsyncNats(httpAsync.Entrypoint, "")
		an.SubscribeAsync(httpAsync.HttpFn)
		logger.Println("[http-Async-Nats] " + httpAsync.Entrypoint + " at " + an.getSubj())
	}

	// start server
	httpSrv = &http.Server{Addr: ":" + strings.TrimPrefix(port, ":"), Handler: g}

	go func() {
		fmt.Println("HTTP started on " + httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("HTTP", err.Error())
		}
	}()
}

func wrapAsync(fn HttpAsync) HttpFn {
	return func(w http.ResponseWriter, r *http.Request) {
		callback := r.Header.Get("X-Callback-Url")
		if callback == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("missing header, X-Callback-Url"))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("Will respond at " + callback + "\n"))
		go runAsyncFn(fn.HttpFn, callback, r)
	}
}

func runAsyncFn(fn HttpFn, callback string, r *http.Request) {
	ww := httptest.NewRecorder()
	fn(ww, r)
	_, err := http.Post(callback, "application/json", ww.Result().Body)
	if err != nil {
		log.Println(err.Error())
	}
}
func Stop() {
	provider.Close()
	cx, can := context.WithTimeout(context.Background(), 2*time.Second)
	defer can()
	if err := httpSrv.Shutdown(cx); err != nil {
		logger.Println("Http-Shutdown " + err.Error())
	} else {
		logger.Println("Http-Shutdown complete")
	}
}
