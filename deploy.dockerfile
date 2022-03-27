
FROM ishan27g/ryo-faas:rfa-deploy-base.v0.1 as build
WORKDIR /app
COPY . .

FROM golang:alpine3.13 as intermediate
WORKDIR /app
COPY --from=build /app/ .
WORKDIR deployments/tmp
RUN go build

FROM alpine:3.13
WORKDIR /app
COPY --from=intermediate /app/deployments/tmp .

EXPOSE 6000
ENTRYPOINT ["./tmp" , "--port", "6000"]



