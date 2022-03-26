FROM golang:alpine3.13 as build

WORKDIR /app
COPY . .

#RUN go mod tidy
#RUN go mod vendor

FROM alpine:3.13
COPY --from=build /app /app
COPY --from=build /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

