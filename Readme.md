# RunYourOwn-FaaS

Functions as a service and json datastore.

- Run `http functions` as individual services (see examples/method2)
- Run `Async / background http` functions over Nats (see examples/async)
- Run functions triggered on changes to the `Json datastore` like `new`,`updated`,`deleted`, `get` (see examples/database-events)
- Run a `combination` of above as a service (examples/database-events)
- Observable functions with built-in [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) tracing

## Definition

#### Define a golang function - `hello/helloWorld.go`

```go
package hello

import (
  "fmt"
  "net/http"
)
// HelloWorld is the function to be deployed
// Note - function should be exported & have correct params
func HelloWorld(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, "Hello, World!\n")
}
```

#### Create its definition File - `deploy.json`

```json
{
  "deploy": [
    {
      "name" : "HelloWorld",
      "filePath": "./example/hello/helloWorld.go"
    }
  ]
}
```

#### Setup / Install ryo-Faas
```shell
# Creates a directory - $HOME/.ry-faas/
# and pulls relevant docker images 
./proxyCli init
```

Will take a few minutes to download the following docker images

- [proxy](https://hub.docker.com/repository/docker/ishan27g/ryo-faas) running at `localhost:9999`
- [database](https://hub.docker.com/repository/docker/ishan27g/ryo-faas) running at `localhost:5000/5001`
- [functionBase](https://hub.docker.com/repository/docker/ishan27g/ryo-faas) attached to internal docker network

- nats:alpine3.15 running at `localhost:4222/8222`
- openzipkin/zipkin:2.23.15 running at `localhost:9411`

#### Start ryo-Faas

```shell
# Ensure docker is running
./proxyCli startFaas
```

#### Deploy

```shell
./proxyCli deploy deploy.json
```
__The function is made available via the proxy at `http://localhost:9999/functions/helloworld`__

Trigger the endpoint and view the traces collected by the default exporter - `Jaeger` running at `http://localhost:16686`
```shell
curl http://localhost:9999/functions/helloworld
open http://localhost:16686
```

## Individual HTTP/ASYNC Functions

__Add flag `--async` to deploy as an `async` function__. See [Async](#####Async Http)

```shell
./proxyCli deploy --async deploy.json
```

## DataStore Event Triggers , or a combination with HTTP/ASYNC Functions

__Add `--main` to deploy a combination of `http` `async` & `events`__
```shell
./proxyCli deploy --main deployMain.json
```

Should export a single `Init()` method that registers the requires triggers, http & async functions. See (example/database-events/)

```go
// NOTE THE PACKAGE NAME, IT SHOULD NOT BE A MAIN PACKAGE
package notMain

import (
 "fmt"
 "net/http"

 FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
 "github.com/Ishan27g/ryo-Faas/store"
)
func HttpMethod(w http.ResponseWriter, r *http.Request) {
 w.WriteHeader(http.StatusAccepted)
 fmt.Fprint(w, "Accepted at method - HttpMethod ..."+"\n")
}
func HttpAsyncMethod(w http.ResponseWriter, r *http.Request) {
 w.WriteHeader(http.StatusAccepted)
 fmt.Fprint(w, "Accepted at method - HttpAsyncMethod..."+"\n")
}
func documentTrigger(document store.Doc) {
 fmt.Println(document.CreatedAt + " " + document.Id + " ---- at GenericCb()")
}
func main() {

// Http

 // register a http method
 FuncFw.Export.Http("HttpMethod", "/method1", HttpMethod)

//Async

 // register your http async method over Nats
 FuncFw.Export.NatsAsync("HttpAsyncMethod-Nats", "/asyncNats", HttpAsyncMethod)
 
 // or register your http async method over Http
 FuncFw.Export.HttpAsync("HttpAsyncMethod", "/async", HttpAsyncMethod)

//DataStore events

     // register a function to be called when a new `payments` document is created
 FuncFw.Export.EventsFor("payments").On(store.DocumentCREATE, documentTrigger)
    // register a function to be called when some existing `bills` document is updated
 FuncFw.Export.EventsFor("bills").On(store.DocumentUPDATE, documentTrigger)
     // register a function to be called when a known `payments` document (by its ID) is retrieved
 FuncFw.Export.EventsFor("payments").OnIds(store.DocumentGET, documentTrigger, "some-known-id")
     // register a function to be called when a known `bills` document (by its ID) is retrieved
 FuncFw.Export.EventsFor("bills").OnIds(store.DocumentGET, documentTrigger, "some-known-id")
}
```

```json
{
  "deploy": [
    {
      "name" : "Database-events",
      "filePath": "/Users/ishan/Desktop/multi/database-events/main.go"
    }
  ]
}
```

#### Get details of all deployments
```shell
./proxyCli details
```
#### Stop
```shell
# stop a function, optional --prune flag
./proxyCli stop [functionName]

# or, stop ryo Faas
./proxyCli stopFaas

# optionally prune all images
./proxyCli prune
```

## How it works

#####Http
functions are run in a manner similar to Google's [functions-framework-go](https://github.com/GoogleCloudPlatform/functions-framework-go).
It simply registers the http-functions and then starts an HTTP server serving that function. (not considering cloudEvents).

#####Async-Http 
functions are run in a manner similar to [OpenFaas](https://docs.openfaas.com/reference/async/). 
The incoming request is serialised and sent to Nats allowing immediate response for the request. 
The Nats message is received, deserialized into the http request and then acted upon. 
The result is sent to a `X-Callback-Url` that is expected in the original request.

#####Store
The `store` publishes `events` to Nats on each `CRUD` operation to the database, allowing subscribers to act on relevant changes

- The function to be deployed along with its directory are copied to `$HOME/.ry-faas/deployments/tmp/`. Using the `ast`  package, the `cli`
  - Verifies the signature
    - of the exported `http-function`, or
    - of the exported `main-service`
  - Generates a new `exported_{function}.go` file (based on this [template](https://github.com/Ishan27g/ryo-Faas/blob/main/pkg/template/template.go)) that registers the provided function with the framework before starting an Http server.
- The generated `service` is then built into a Docker image and run as its own container

