FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/iam-service /iam-service

EXPOSE 8081

USER 65534

ENTRYPOINT ["/iam-service"]
