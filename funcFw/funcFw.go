package FuncFw

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	database "github.com/Ishan27g/ryo-Faas/database/client"
	"github.com/Ishan27g/ryo-Faas/examples/plugins"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	databaseAddress              = "localhost:5000"
	port                         = ""
	jp                           = plugins.InitJaeger(context.Background(), "ryo-Faas-agent", "", "http://jaeger:14268/api/traces")
	httpSrv         *http.Server = nil
	logger                       = log.New(os.Stdout, "func-fw", log.LstdFlags)
	healthCheckUrl               = "/healthcheck"
	stopUrl                      = "/stop"
	Export                       = funcFw{
		httpFns:       make(map[string]*HttpFunction),
		httpAsync:     make(map[string]*HttpAsync),
		httpAsyncNats: make(map[string]*HttpAsync),
		storeEvents:   nil,
	}
)

func Start(port string) {
	// apply store event handlers
	if Export.storeEvents != nil {
		if database.Connect(databaseAddress) == nil {
			log.Fatal("Store: Cannot connect to database")
		}
		if !ApplyEvents() {
			log.Fatal("Store: Unable to apply event cbs")
		}
	}

	// apply http handlers
	for entrypoint, function := range Export.GetHttp() {
		otelHandler := otelhttp.NewHandler(http.HandlerFunc(wrapHttp(function.HttpFn)), "deployed-service-"+entrypoint)
		http.Handle(function.UrlPath, otelHandler)
		logger.Println("[http] " + function.Entrypoint + " at " + function.UrlPath)
	}

	// apply http async handlers
	for entrypoint, httpAsync := range Export.GetHttpAsync() {
		otelHandler := otelhttp.NewHandler(http.HandlerFunc(wrapAsync(httpAsync)), "deployed-service-async-"+entrypoint)
		http.Handle(httpAsync.UrlPath, otelHandler)
		logger.Println("[http-Async] " + httpAsync.Entrypoint + " at " + httpAsync.UrlPath)
	}

	// apply http async nats handlers
	for _, httpAsync := range Export.GetHttpAsyncNats() {
		an := NewAsyncNats(httpAsync.Entrypoint, "")
		an.SubscribeAsync(httpAsync.HttpFn)
		logger.Println("[http-Async-Nats] " + httpAsync.Entrypoint + " at " + an.getSubj())
	}

	// healthcheck
	http.Handle(healthCheckUrl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	http.Handle(stopUrl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Stop()
		w.WriteHeader(http.StatusOK)
	}))

	httpSrv = &http.Server{Addr: ":" + port}
	logger.Println("HTTP listening on " + httpSrv.Addr)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Println("HTTP-Error", err.Error())
		}
	}()
}

func wrapHttp(fn HttpFn) HttpFn {
	return func(writer http.ResponseWriter, request *http.Request) {
		fn(writer, request)
	}
}

func wrapAsync(httpAsync *HttpAsync) HttpFn {
	return func(w http.ResponseWriter, r *http.Request) {
		callback := r.Header.Get("X-Callback-Url")
		if callback == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing header, X-Callback-Url"))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Will respond at " + callback + "\n"))
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
	jp.Close()
	cx, can := context.WithTimeout(context.Background(), 2*time.Second)
	defer can()
	if err := httpSrv.Shutdown(cx); err != nil {
		logger.Println("Http-Shutdown " + err.Error())
	} else {
		logger.Println("Http-Shutdown complete")
	}
}
