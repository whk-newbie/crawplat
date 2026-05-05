FROM scratch

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine:latest /usr/share/zoneinfo /usr/share/zoneinfo

COPY .docker-bin/project-service /project-service

EXPOSE 8082

USER 65534

ENTRYPOINT ["/project-service"]
