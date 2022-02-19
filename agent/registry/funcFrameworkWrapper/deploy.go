package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

var deploy = func(name string, fn interface{}) {

	fnHTTP, ok := fn.(func(http.ResponseWriter, *http.Request))
	if !ok {
		panic("expected function to have signature func(listen.ResponseWriter, *listen.Request)")
	}

	ctx := context.Background()
	url := "/" + os.Getenv("URL")
	fmt.Println("deploying at ", url)
	if err := funcframework.RegisterHTTPFunctionContext(ctx, url, fnHTTP); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}

	port := os.Getenv("PORT")
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
