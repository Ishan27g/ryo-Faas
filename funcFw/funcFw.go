package FuncFw

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ishan27g/ryo-Faas/plugins"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	port                 = ""
	jp                   = plugins.InitJaeger(context.Background(), "ryo-Faas-agent", "", "http://jaeger:14268/api/traces")
	httpSrv *http.Server = nil
	logger               = log.New(os.Stdout, "func-fw", log.LstdFlags)
)

func Start(port string) {
	for entrypoint, function := range Export.Get() {
		otelHandler := otelhttp.NewHandler(http.HandlerFunc(function.HttpFn), "deployed-service-"+entrypoint)
		http.Handle(function.UrlPath, otelHandler)
		logger.Println(function.Entrypoint + " at " + function.UrlPath)
	}
	httpSrv = &http.Server{Addr: ":" + port}
	logger.Println("HTTP listening on " + httpSrv.Addr)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Println("HTTP-Error", err.Error())
		}
	}()
}
func Stop() {
	jp.Close()
	cx, can := context.WithTimeout(context.Background(), 2*time.Second)
	defer can()
	if err := httpSrv.Shutdown(cx); err != nil {
		logger.Println("Http-Shutdown " + err.Error())
	} else {
		logger.Println(err.Error())
	}
}
