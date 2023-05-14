#!/usr/bin/env bash

# test scaling

task build
go build -o ryo-Faas cli.go
./ryo-Faas sto
./ryo-Faas i
./ryo-Faas sta

sleep 3
./ryo-Faas deploy examples/deployHello.json
./ryo-Faas deploy pkg/scale/deploy-scale.json
sleep 3

for i in {1..100} ; do
    curl http://localhost:9999/functions/hello
    sleep 0.1
done

echo docker container ls | grep -c rfa-deploy-hello
exit