FROM golang:1.24 AS build

WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go-echo main.go

FROM gcr.io/distroless/static-debian12
COPY --from=build /go-echo /go-echo
ENTRYPOINT ["/go-echo"]
