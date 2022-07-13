package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var defaultHttp = ":5999"
var httpPort = flag.String("http", defaultHttp, "http port")

func main() {
	provider := tracing.Init("jaeger", "async-server", "")
	defer provider.Close()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(otelgin.Middleware("async-server"))
	r.Any("/any", func(c *gin.Context) {
		fmt.Println(c.Request.URL.RequestURI(), " from ", c.Request.RemoteAddr)
		span := trace.SpanFromContext(c.Request.Context())
		span.SetAttributes(attribute.Key("okokok").String("ok"))
		body, _ := ioutil.ReadAll(c.Request.Body)
		println("Response - ", string(body))
		c.JSON(http.StatusOK, nil)
	})
	fmt.Println("Listening for /any on", defaultHttp)
	log.Fatalf(r.Run(*httpPort).Error())
}
