FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

#RUN go mod tidy
#RUN go mod vendor

#ENV PORT_START=5000
#ENV NUM_PORTS=5

#EXPOSE 9000
WORKDIR proxy
RUN go build main.go

FROM alpine:3.13
WORKDIR /app
COPY --from=build /app/proxy/main /app/

ENTRYPOINT ["./main"]