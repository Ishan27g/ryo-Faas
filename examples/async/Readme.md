A long-running function/http-handler, deployed as an `async` function.

If the incoming request contains the header `X-Callback-Url`, the response of the `async` function is `Posted` to this url

The `external-server` is a simple server that listens for the response sent by the `async` function. It runs on port 5999.
It receives the response sent back by the `async` function, analogous to a `webhook`

After deploying the example, sending the below request will trigger the `external-server` once the `async` method has completed

```shell
# host.docker.internal -> host's localhost (on mac) 
curl -X POST http://localhost:9999/functions/methodasync -H 'X-Callback-Url:http://host.docker.internal:5999/any'
```
