A handler that sends a http request to another deployed handler [method2](https://github.com/Ishan27g/ryo-Faas/tree/main/examples/method2) 

The address of the secondary service is passed as a json payload.

```shell
curl -X POST http://localhost:9999/functions/method1 -H 'Content-Type: application/json' -d '{                    
    "data": "http://host.docker.internal:9999/functions/method2"
  }'
```