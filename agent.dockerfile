FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

#RUN go mod tidy
#RUN go mod vendor

#ENV PORT_START=5000
#ENV NUM_PORTS=5
#EXPOSE 9000 5000 5001 5002 5003 5004

EXPOSE 9000
#ENTRYPOINT ["go", "run", "agent/main.go"]