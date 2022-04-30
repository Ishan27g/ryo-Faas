#!/usr/bin/env bash

task imgProxy
go run cli.go sto
rm -rf /Users/ishan/Documents/Drive/golang/ryo-Faas/
go run cli.go init
go run cli.go sta
go run cli.go deploy examples/deploy-otel.json