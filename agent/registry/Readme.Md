
# Deployed Function

Deploying `examples/method1` and `examples/method2`

- uploads the respective directories & generates the `init()` definition for `template/template.go`

- at  `agent/registry/deploy/functions`

- `agent` starts funcFw on `env:PORT`

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    method1 "github.com/Ishan27g/ryo-Faas/agent/registry/deploy/functions/method1"
    method2 "github.com/Ishan27g/ryo-Faas/agent/registry/deploy/functions/method2"
    FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

// init definition gets generated
func init() {
	// http 
    FuncFw.Export.Http("Method1", "/method1", method1.Method1)
	// async http
	FuncFw.Export.HttpAsync("Async", "/async", method2.Method2)
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {return }

    FuncFw.Start(port)

    closeLogs := make(chan os.Signal, 1)
    signal.Notify(closeLogs, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    <-closeLogs

    FuncFw.Stop()
}
```
