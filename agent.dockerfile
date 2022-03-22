FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

#RUN go mod tidy
#RUN go mod vendor

#ENV PORT_START=5000
#ENV NUM_PORTS=5
EXPOSE 6000 6001 6002 6003 6004

#EXPOSE 9000
ENTRYPOINT ["go", "run", "agent/main.go"]