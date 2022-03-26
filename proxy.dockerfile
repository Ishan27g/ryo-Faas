FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

WORKDIR proxy
RUN go build main.go

FROM alpine:3.13
WORKDIR /app
COPY --from=build /app/proxy/main /app/

ENTRYPOINT ["./main"]