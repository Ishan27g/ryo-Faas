# RunYourOwn-FaaS

Run your own `Functions as a service` and `Json datastore` with built in OpenTelemetry tracing

- Run `http functions` as individual services (see examples/method2)
- Run `Async / background http` functions over Nats (see examples/async)
- Run functions triggered on changes to the `Json datastore` like `new`,`updated`,`deleted`, `get` (see examples/database-events)
- Run a `combination` of above as a service (examples/database-events)

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
      "packageDir": "./example/hello",
      "name" : "HelloWorld",
      "filePath": "./example/hello/helloWorld.go",
    }
  ]
}
```

#### Start ryo-Faas

Will take a few minutes to download the following docker images

- [proxy](https://hub.docker.com/repository/docker/ishan27g/ryo-faas) running at `localhost:9999`
- [database](https://hub.docker.com/repository/docker/ishan27g/ryo-faas) running at `localhost:5000/5001`
- [functionBase](https://hub.docker.com/repository/docker/ishan27g/ryo-faas) attached to internal docker network

- nats:alpine3.15	running at `localhost:4222/8222`
- openzipkin/zipkin:2.23.15 running at `localhost:9411`

```shell
./proxyCli startFaas
```

#### Deploy

```shell
./proxyCli deploy deploy.json
```

- The function `HelloWorld` gets deployed as its own Docker container behind a network
- Proxy forwards the corresponding requests to the `HelloWorld` running at `http://rfa-helloworld:6000`

#### Get details of all deployments

```shell
./proxyCli details
```
__The function is available via the proxy at `http://localhost:9999/functions/helloworld`__

## Individual HTTP/ASYNC Functions

Should follow the standard go http library for the handler
```go
func SomeHandler(w http.ResponseWriter, r *http.Request){}
```
> __Add `"async": true` to deploy as an `async` function__ ()

```json
{
  "deploy": [
    {
      "packageDir": "/Users/ishan/Desktop/multi/async",
      "name" : "MethodAsync",
      "filePath": "/Users/ishan/Desktop/multi/async/main.go",
      "async": true
    }
  ]
}
```

## DataStore Event Triggers , or a combination with HTTP/ASYNC Functions

Should export a single `Init()` method that registers the requires triggers, http & async functions. See (example/database-events/)

```go
# NOTE THE PACKAGE NAME, IT SHOULD NOT BE A MAIN PACKAGE
package notMain

import (
	"fmt"
	"net/http"
	"time"

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
	FuncFw.Export.Async("HttpAsyncMethod", "/async", HttpAsyncMethod)

//DataStore events

    	// register a function to be called when a new `payments` document is created
	FuncFw.Export.EventsFor("payments").On(store.DocumentCREATE, documentTrigger)
   	// register a function to be called when some existing `bills` document is updated
	FuncFw.Export.EventsFor("bills").On(store.DocumentUPDATE, documentTrigger)
    	// register a function to be called when a known `payments` document (by its ID) is retrieved
	FuncFw.Export.EventsFor("payments").OnIds(store.DocumentGET, cb, "some-known-id")
    	// register a function to be called when a known `bills` document (by its ID) is retrieved
	FuncFw.Export.EventsFor("bills").OnIds(store.DocumentGET, cb, "some-known-id")
}
```

> __Add `"mainProcess" : true` to deploy a combination of `http` `async` & `events`__

```json
{
  "deploy": [
    {
      "mainProcess" : true,
      "packageDir": "/Users/ishan/Desktop/multi/database-events",
      "name" : "Database-events",
      "filePath": "/Users/ishan/Desktop/multi/database-events/main.go"
    }
  ]
}
```

## How it works

- `Http` functions are run in a manner similar to Google's [functions-framework-go](https://github.com/GoogleCloudPlatform/functions-framework-go).
It simply registers the http-functions and then starts an HTTP server serving that function. (not considering cloudEvents).

- `Async Http` functions are run in a manner similar to [OpenFaas](https://docs.openfaas.com/reference/async/). The incoming request is serialised and sent to Nats allowing immediate response for the request. The Nats message is received, deserialised into the http request and then acted upon. The result is sent to a `X-Callback-Url` that is expected in the original request.

- The `store` publishes `events` to Nats on each `CRUD` operation to the database, allowing subscribers to act on relevant changes

- The function to be deployed along with its directory are copied to `./deployments/tmp/`. Using the `ast`  package, the `cli`
  - verifies the signature
    - of the exported `http-function`, or
    - of the exported `main-service`
  - generates a `exported_{function}.go` file which registers the provided function with the `functionFramework` (https://github.com/Ishan27g/ryo-faas/blob/main/remote/agent/funcFrameworkWrapper/template.go) before starting an http-server. See [template.go]
- The generated `service` is then built into a Docker image and run as its own container

