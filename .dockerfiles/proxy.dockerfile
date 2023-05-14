FROM golang:alpine3.16 as build

WORKDIR /app
COPY . .

WORKDIR proxy
RUN go build main.go

FROM alpine:3.16
WORKDIR /app
COPY --from=build /app /app

ENTRYPOINT ["./proxy/main"]