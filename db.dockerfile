FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

RUN go build database/main.go

FROM alpine:3.13
WORKDIR /app
COPY --from=build /app/main /app/

ENTRYPOINT ["./main"]