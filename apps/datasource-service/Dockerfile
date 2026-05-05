FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/datasource-service /datasource-service

EXPOSE 8086

USER 65534

ENTRYPOINT ["/datasource-service"]
