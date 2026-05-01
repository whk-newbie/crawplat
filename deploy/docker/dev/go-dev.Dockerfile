FROM golang:1.25-alpine

ENV GOPROXY=https://goproxy.cn,direct
ENV PATH=/usr/local/go/bin:/go/bin:${PATH}

RUN apk add --no-cache bash docker-cli git && \
    (go install github.com/air-verse/air@v1.61.7 || \
     GOPROXY=https://goproxy.cn,direct go install github.com/air-verse/air@v1.61.7)

WORKDIR /workspace
