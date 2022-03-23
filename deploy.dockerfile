
FROM ishan27g/ryo-faas:rfa-deploy-base.v0.1 as build
WORKDIR /app
COPY deployments .

ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR deployments
RUN go build

FROM alpine:3.13
WORKDIR /app
COPY --from=build /app/deployments .

EXPOSE 6000
ENTRYPOINT ["./deployments", "--port", "6000" ]



