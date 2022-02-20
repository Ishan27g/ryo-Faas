rm -r agent/registry/deploy/functions
function Deploy() {
  cd cli
  go run proxyCli.go add localhost:9000
  go run proxyCli.go r deploy.json
  cd ..
}
Deploy
cd testClient
Max=10
sleep 3
for i in $(seq 2 $Max)
do
  go run client.go &&  curl -X POST http://localhost:9002/functions/method1 -H 'Content-Type: application/json' -d '{
      "data": "http://host.docker.internal:9002/functions/method2"
    }'
done


