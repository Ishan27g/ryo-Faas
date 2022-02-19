# RunYourOwn-FaaS

- Run agents on your infrastructure
- Define function definitions and deploy them as services on agents

## Agent

Deploys function as its own process on the agent machine. Runs on separate machines (ideally).

`Start cloud-agents on various machines`

```shell
make remote && ./agent -h 
./agent
# HTTP started on :9000`
```

`Note` : Agent binary should be run from the root of the repository (or via docker) as the generated file definition needs to be run from a valid package.

## Proxy

Communicates with agents to deploy functions.

- Acts as a reverse proxy for calling functions deployed on agents.

`Run proxy on another machine`

```shell
make proxy
./proxy
# HTTP started on :9002 
# GRPC server started on [::]:9001
```

## Definition

`Define a golang function`

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

`Create its definition File`

- Deploy a function to agents - `deploy.json`

```json
{
  "deploy": [
    {
      "packageDir": "./example/hello",
      "name" : "HelloWorld",
      "filePath": "./example/hello/helloWorld.go",
      "agentAddr" : ["localhost:9000"]
    }
  ]
}
```

- List functions that are deployed at agents - `list.json`

```json
{
  "list": {
    "agentAddr" : ["http://localhost:9000"]
  }
}
```

## Deploy

### Via CLI

The `cli` communicates with the `proxy` . Can also be configured to communicate with the `agent` directly. [link to it below]

```shell
make cli && ./proxyCli -help

# deploy the fn
./proxyCli def deploy.json
> name:"HelloWorld"  url:"http://localhost:5015/helloworld"  status:"Deploying"

# list deployed fn
./proxyCli def list.json
> name:"HelloWorld"  url:"http://localhost:5015/helloworld"  status:"Deployed"

# call the deployed fn
curl http://localhost:5015/helloworld
> Hello, World!
```

### Via Http via `curl,postman` ...

##### Deploy `HelloWorld` on `agent` running at `localhost:9000`

```shell
# Request to proxy
curl --location --request POST 'http://Ishans-MacBook-Pro.local:9002/deploy' \
--header 'Content-Type: application/json' \
--data-raw '[
    {
        "dir": "/Users/ishan/go/src/github.com/Ishan27g/ryo-faas/example/hello",
        "name": "HelloWorld",
        "filePath": "/Users/ishan/go/src/github.com/Ishan27g/ryo-faas/example/hello/helloWorld.go",
        "toAgent": "localhost:9000"
    }
]' 
# Response from proxy
[{"name": "HelloWorld","status": "Deploying","url": "http://localhost:9002/functions/helloworld","proxy": "http://localhost:5003","atAgent": "localhost:9000"}]
```
- The function `HelloWorld` gets deployed on the `agent` at `localhost:9000`.
- The function is available via the proxy at `http://localhost:9002/functions/helloworld`
- Proxy forwards the request to the `service` running on the `agent` at `http://localhost:5003`

##### Get details of all deployments
```shell
# Request to proxy
curl --location --request GET 'http://Ishans-MacBook-Pro.local:9002/details'
# Response from proxy
[{"name": "HelloWorld","status": "Deployed","url": "http://localhost:9002/functions/helloworld","proxy": "http://localhost:5003","atAgent": "localhost:9000"}]
```

##### Get details of deployments at `agent` running at `localhost:9000`
```shell
# Request to proxy
curl --location --request POST 'http://Ishans-MacBook-Pro.local:9002/list' \
--header 'Content-Type: application/json' \
--data-raw '{
    "atAgent":"localhost:9000"
}'
# Response from proxy
[{"name": "HelloWorld","status": "Deployed","url": "http://localhost:9002/functions/helloworld","proxy": "http://localhost:5003","atAgent": "localhost:9000"}]
```

##### Get the logs for `HelloWorld`

```shell
# Request to proxy
curl --location --request GET 'http://Ishans-MacBook-Pro.local:9002/log?entrypoint=HelloWorld'
# Response from proxy
{
    "function": {"name": "HelloWorld","status": "Deployed","url": "http://localhost:9002/functions/helloworld","proxy": "http://localhost:5003","atAgent": "localhost:9000"},
    "logs": "deploying at /helloworld"
}
```

### Using `client` interface for the `agent`

`agent` provides a client-interface to communicate with it.

- The `proxy` uses this interface to communicate with the `agent`.
- The `cli` uses this interface to communicate with the `proxy`.

###### `proxy` implements all `agent` methods which makes communication between the client, proxy and agent homogenous. This allows the `cli` or any client to effectivelly bypass the proxy and communicate directly with the agent.

- In case of a single `agent`, the proxy only adds overhead and thus can be skipped.
- With multiple `agents`, the proxy can act as a common gateway & reverse proxy for the agents.
```go
package transport
type Interface interface {
	Deploy(w ...proto.FunctionJson) []proto.FunctionJsonRsp
	Stop(fnName string) proto.FunctionJsonRsp
	List(empty *proto.Empty) []proto.FunctionJsonRsp
	Log(fnName string) *proto.FunctionLogs
	Details() []proto.FunctionJsonRsp
}
```
```go

package main

import "github.com/Ishan27g/ryo-faas/transport"

...

// same interface to talk to proxy or agent
var cli transport.Interface

// address of the proxy or agent
var clientAdr = ":9000"

func main() {
	// connect to proxy or agent
	// via rpc
	cli = transport.NewAgent(transport.RPC, clientAdr)
	// OR via http
	cli = transport.NewAgent(transport.HTTP, clientAdr)

	cli.Deploy(...)
	cli.List(...)
	cli.Stop(...)
	cli.Log(...)
	cli.Details()
}
```

### How it works

Google's [functions-framework-go](https://github.com/GoogleCloudPlatform/functions-framework-go) is used to deploy a function as its own service.
It simply registers the http-functions and then starts an HTTP server serving that function.

- The function to be deployed along with its directory are compressed and uploaded to the `agent`.
- Using the `ast`  package, the `agent`
    - verifies the signature of the exported function.
    - generates a `main_{function}.go` file which registers the provided function
      with the `cloudFunctionFramework`.  See [template.go](https://github.com/Ishan27g/ryo-faas/blob/main/remote/agent/funcFrameworkWrapper/template.go) & [deploy.go](https://github.com/Ishan27g/ryo-faas/blob/main/remote/agent/funcFrameworkWrapper/deploy.go)
- The generated `service` is then run as a new system process on the `agent`'s machine.

This function is then exposed via the `proxy` which routes the request to its corresponding `service`
