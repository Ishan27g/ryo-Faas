#!/usr/bin/env bash

if [ -z "$1" ]; then
  echo "NOT BUILDING IMAGES FOR TEST!!!!!!"
else
  task build
fi

# test as ryo-faas `command`, without taskfile env
go build -o ryo-Faas cli.go
./ryo-Faas sto
./ryo-Faas i
./ryo-Faas sta
./ryo-Faas deploy --main examples/deployMain.json
curl http://localhost:9999/functions/database-events/pay
curl http://localhost:9999/functions/database-events/pay
curl http://localhost:9999/functions/database-events/pay
curl http://localhost:9999/functions/database-events/pay
curl http://localhost:9999/functions/database-events/get