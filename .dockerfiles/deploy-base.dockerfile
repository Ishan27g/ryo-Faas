FROM golang:alpine3.16 as build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go mod vendor

FROM alpine:3.16
COPY --from=build /app /app
#COPY --from=build /usr/local/go/ /usr/local/go/
#ENV PATH="/usr/local/go/bin:${PATH}"

