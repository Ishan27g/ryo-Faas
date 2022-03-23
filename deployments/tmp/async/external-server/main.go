package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var defaultHttp = ":5999"
var httpPort = flag.String("http", defaultHttp, "http port")

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Any("/any", func(c *gin.Context) {
		fmt.Println(c.Request.URL.RequestURI(), " from ", c.Request.RemoteAddr)
		fmt.Println(c.Request.Header)
		fmt.Println(c.Request.Host)
		body, _ := ioutil.ReadAll(c.Request.Body)
		println("Response - ", string(body))
		c.JSON(http.StatusOK, nil)
	})
	fmt.Println("Listening for /any on", defaultHttp)
	log.Fatalf(r.Run(*httpPort).Error())
}
