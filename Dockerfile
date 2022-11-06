FROM golang:1.17-alpine as builder

RUN apk update && apk add --no-cache git bash

ADD . /app
WORKDIR /app

RUN go get ./...
RUN go build -o application

FROM alpine:latest

COPY --from=builder /app/application /application

EXPOSE 8080
ENTRYPOINT ["/application"]