FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/monitor-service /monitor-service

EXPOSE 8087

USER 65534

ENTRYPOINT ["/monitor-service"]
