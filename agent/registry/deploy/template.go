package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ishan27g/ryo-Faas/plugins"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var handlerFunc func(w http.ResponseWriter, r *http.Request)
var entrypoint string

// init definition gets generated to call deploy()
func init() {

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		return
	}
	url := "/" + os.Getenv("URL")
	fmt.Println("deploying at ", url)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jp := plugins.InitJaeger(ctx, "ryo-Faas-agent", "deployed-service-"+entrypoint, "http://jaeger:14268/api/traces") //match with docker hostname
	defer jp.Close()
	// _ = jp.Tracer("function-with-otel")

	// https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/net/http/otelhttp/example
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(handlerFunc), "deployed-service-"+entrypoint)
	http.Handle(url, otelHandler)

	httpSrv := &http.Server{
		Addr: ":" + port,
	}
	go func() {
		fmt.Println("HTTP started on " + httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("HTTP", err.Error())
		}
	}()
	<-catchExit()
	fmt.Println("EXITING?")
	cx, can := context.WithTimeout(context.Background(), 2*time.Second)
	defer can()
	if err := httpSrv.Shutdown(cx); err != nil {
		fmt.Println("Http-Shutdown " + err.Error())
	} else {
		fmt.Println(err.Error())
	}
	//if err := http.ListenAndServe(":"+port, nil); err != nil {
	//	log.Fatalf("httpListenAndServe: %v\n", err)
	//}
}
func catchExit() chan bool {
	stop := make(chan bool, 1)
	closeLogs := make(chan os.Signal, 1)
	signal.Notify(closeLogs, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-closeLogs
		stop <- true
	}()
	return stop
}
