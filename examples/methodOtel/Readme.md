An example of extracting OpenTelemetry span from a client request and updating it.

Ideally, another service/client would create the traces that are expected by this handler.

[exampleOtelClient](https://github.com/Ishan27g/ryo-Faas/blob/main/examples/methodOtel/exampleOtelClient.go) is a simple http-client that sends a request with its `span` and added `baggage` .

The traces can be viewed via `zipkin` or `jaeger`
