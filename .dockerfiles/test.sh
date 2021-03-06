#!/usr/bin/env bash

if [ -z "$1" ]; then
  echo ""
else
  task imgProxy
fi

go run cli.go sto
rm -rf /Users/ishan/Documents/Drive/golang/ryo-Faas/
rm -rf /Users/ishan/Desktop/multi/*
cp -r examples/* /Users/ishan/Desktop/multi/
go run cli.go init
go run cli.go sta

go run cli.go deploy pkg/scale/deploy-scale.json
go run cli.go deploy examples/deploy-otel.json

#go run cli.go deploy examples/deploy-otel.json
#go run cli.go deploy examples/deploy-otel.json