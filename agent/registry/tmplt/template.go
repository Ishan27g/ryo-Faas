package z

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ishan27g/ryo-Faas/agent/registry/deploy"
	"github.com/Ishan27g/ryo-Faas/plugins"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// init definition gets generated
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

	jp := plugins.InitJaeger(ctx, "ryo-Faas-agent", "", "http://jaeger:14268/api/traces") //match with docker hostname
	defer jp.Close()
	for entrypoint, function := range FuncFw.Get() {
		otelHandler := otelhttp.NewHandler(http.HandlerFunc(function.HttpFn), "deployed-service-"+entrypoint)
		http.Handle(function.UrlPath, otelHandler)
	}
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
