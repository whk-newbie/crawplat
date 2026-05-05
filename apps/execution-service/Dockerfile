FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/execution-service /execution-service

EXPOSE 8085

USER 65534

ENTRYPOINT ["/execution-service"]
