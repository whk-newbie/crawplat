FROM golang:1.24-alpine

RUN apk add --no-cache bash docker-cli git && \
    go install github.com/air-verse/air@v1.61.7

WORKDIR /workspace
