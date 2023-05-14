FROM golang:alpine3.16 as build

WORKDIR /app
COPY . .

RUN go build database/main.go

FROM alpine:3.16
WORKDIR /app
COPY --from=build /app/main /app/

ENTRYPOINT ["./main"]