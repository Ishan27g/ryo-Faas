FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

#RUN go mod tidy
#RUN go mod vendor

RUN go build database/main.go
#ENV PORT_START=5000
#ENV NUM_PORTS=5
#EXPOSE 9000 5000 5001 5002 5003 5004

FROM alpine:3.13
WORKDIR /app
COPY --from=build /app/main /app/

EXPOSE 5001
ENTRYPOINT ["./main"]