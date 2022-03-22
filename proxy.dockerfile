FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

#RUN go mod tidy
#RUN go mod vendor

#ENV PORT_START=5000
#ENV NUM_PORTS=5

#EXPOSE 9000

ENTRYPOINT ["go", "run", "proxy/main.go"]