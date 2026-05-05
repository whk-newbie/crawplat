FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/spider-service /spider-service

EXPOSE 8083

USER 65534

ENTRYPOINT ["/spider-service"]
