FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/node-service /node-service

EXPOSE 8084

USER 65534

ENTRYPOINT ["/node-service"]
