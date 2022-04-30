#!/usr/bin/env bash

if [ -z "$1" ]; then
  echo ""
else
  task imgProxy
fi

go run cli.go sto
rm -rf /Users/ishan/Documents/Drive/golang/ryo-Faas/
go run cli.go init
chmod 777 /Users/ishan/Documents/Drive/golang/ryo-Faas
go run cli.go sta
go run cli.go deploy examples/deploy-otel.json