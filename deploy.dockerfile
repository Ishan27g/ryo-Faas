
FROM ishan27g/ryo-faas:rfa-deploy-base.v0.1 as build
WORKDIR /app
COPY . .

ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR deployments/tmp
RUN go build

FROM alpine:3.13
WORKDIR /app
COPY --from=build /app/deployments/tmp .

EXPOSE 6000
ENTRYPOINT ["./tmp" , "--port", "6000"]



