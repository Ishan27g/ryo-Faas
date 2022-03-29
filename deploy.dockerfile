FROM golang:alpine3.13 as build
LABEL label=rfa
WORKDIR /app
# copy from baseimage to get pre vendored packages
COPY --from=ishan27g/ryo-faas:rfa-deploy-base.v0.1 /app .
COPY deployments deployments
WORKDIR deployments/tmp
RUN go build

FROM alpine:3.13
LABEL label=rfa
WORKDIR /app
COPY --from=build /app/deployments/tmp .

EXPOSE 6000
ENTRYPOINT ["./tmp" , "--port", "6000"]
