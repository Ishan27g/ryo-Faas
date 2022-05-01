package FuncFw

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	database "github.com/Ishan27g/ryo-Faas/database/client"
	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
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
	}
	provider tracing.TraceProvider
)

func Start(port string) {
	serviceName, _ = os.Hostname()

	gin.SetMode(gin.DebugMode)
	g := gin.New()
	g.Use(gin.Logger())
	g.Use(gin.Recovery())

	if jaegerHost == "" && zipkinHost != "" {
		provider = tracing.Init("zipkin", serviceName, serviceName)
	}
	if zipkinHost == "" && jaegerHost != "" {
		provider = tracing.Init("jaeger", serviceName, serviceName)
	}
	if provider == nil {
		provider = tracing.Init("jaeger", serviceName, serviceName)
	}

	g.Use(otelgin.Middleware(serviceName))

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
		// otelHandler := otelhttp.NewHandler(http.HandlerFunc(function.HttpFn), "deployed-service-"+entrypoint)
		//http.Handle(function.UrlPath, otelHandler)
		//http.Handle(function.UrlPath+"/", otelHandler)
		g.Any(function.UrlPath, gin.WrapH(http.HandlerFunc(function.HttpFn)))
		g.Any(function.UrlPath+"/*any", gin.WrapH(http.HandlerFunc(function.HttpFn)))

		logger.Println("[http] " + function.Entrypoint + " at " + function.UrlPath)
	}

	for _, httpAsync := range Export.getHttpAsync() {
		// otelHandler := otelhttp.NewHandler(http.HandlerFunc(wrapAsync(httpAsync)), "deployed-service-async-"+entrypoint)
		//http.Handle(httpAsync.UrlPath, otelHandler)
		//http.Handle(httpAsync.UrlPath+"/", otelHandler)
		g.Any(httpAsync.UrlPath, gin.WrapH(http.HandlerFunc(wrapAsync(httpAsync))))
		g.Any(httpAsync.UrlPath+"/*any", gin.WrapH(http.HandlerFunc(wrapAsync(httpAsync))))

		logger.Println("[http-Async] " + httpAsync.Entrypoint + " at " + httpAsync.UrlPath)
	}

	// apply http gin handlers
	for _, function := range Export.getHttpGin() {
		// otelHandler := otelhttp.NewHandler(http.HandlerFunc(function.gf), "deployed-service-async-"+entrypoint)
		//http.Handle(httpAsync.UrlPath, otelHandler)
		//http.Handle(httpAsync.UrlPath+"/", otelHandler)
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

	// healthcheck
	//http.Handle(healthCheckUrl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusOK)
	//}))
	g.Any(healthCheckUrl, gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	//http.Handle(stopUrl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	defer Stop()
	//	w.WriteHeader(http.StatusOK)
	//}))
	g.Any(stopUrl, gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Stop()
		w.WriteHeader(http.StatusOK)
	})))

	// start server
	httpSrv = &http.Server{Addr: ":" + port, Handler: g}

	transport.Init(context.Background(), transport.WithHandler(httpSrv.Handler), transport.WithHttpPort(httpSrv.Addr)).
		Start()
}

func wrapAsync(httpAsync *HttpAsync) HttpFn {
	return func(w http.ResponseWriter, r *http.Request) {
		callback := r.Header.Get("X-Callback-Url")
		if callback == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("missing header, X-Callback-Url"))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("Will respond at " + callback + "\n"))
		go runAsyncFn(httpAsync, callback, r)
	}
}

func runAsyncFn(httpAsync *HttpAsync, callback string, r *http.Request) {
	ww := httptest.NewRecorder()
	httpAsync.HttpFn(ww, r)
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
