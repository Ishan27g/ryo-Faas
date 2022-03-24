# RunYourOwn-FaaS

Run your own `Functions as a service` and `Json datastore`.

Deploy `functions` and `triggers` as containers.

- run `http` functions as individual services (examples/method2)
- run `async` / `background` http functions over Nats (examples/async)
- run functions on `Json datastore` triggers like `new-doc`,`updated-doc`, `deleted-doc` (examples/database-events)
- run a `combination` of above as a service (examples/database-events)

- built in opentelemetry `traces` & prometheus `metrics`

## Definition

`Define a golang function` - `hello/helloWorld.go`

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

`Create its definition File` - `deploy.json`

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

__Add `"async": true` to deploy as an `async` function__ ()

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

__Add `"mainProcess" : true` to deploy a combination of `http` `async` & `events`__

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

## Start ryo-Faas

Will take a few minutes to download the docker images

- [proxy] ()
- [database] ()
- [nats] ()
- [functionBase] ()

```shell
./proxyCli startFaas
```

## Deploy

```shell
./proxyCli deploy deploy.json
```

- The function `HelloWorld` gets deployed as its own Docker container behind a network
- The function is available via the proxy at `http://localhost:9999/functions/helloworld`
- Proxy forwards the corresponding requests to the `HelloWorld` running at `http://rfa-helloworld:6000`

## Get details of all deployments

```shell
./proxyCli details
```

### How it works

- `Http` functions are run in a manner similar to Google's [functions-framework-go](https://github.com/GoogleCloudPlatform/functions-framework-go).
It simply registers the http-functions and then starts an HTTP server serving that function. (not considering cloudEvents).

- `Async Http` functions are run in a manner similar to [OpenFaas](https://docs.openfaas.com/reference/async/). The incoming request is serialised and sent to Nats allowing immediate response for the request. The Nats message is received, deserialised into the http request and then acted upon. The result is sent to a `X-Callback-Url` that is expected in the original request.

- The `store` publishes `events` to Nats on each `CRUD` operation to the database, allowing subscribers to act on relevant changes

- The function to be deployed along with its directory are copied to `./deployments/tmp/`. Using the `ast`  package, the `cli`
  - verifies the signature
    - of the exported `http-function`, or
    - of the exported `main-service`
  - generates a `exported_{function}.go` file which registers the provided function with the `functionFramework`.  See [template.go](https://github.com/Ishan27g/ryo-faas/blob/main/remote/agent/funcFrameworkWrapper/template.go) before starting an http-server.
- The generated `service` is then built into a Docker image and run as its own container
