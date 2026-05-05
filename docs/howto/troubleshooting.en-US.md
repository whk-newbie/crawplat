# Troubleshooting

## Service Startup Failures

### Go Service Compilation Errors

**Symptom**: After `make dev-up`, a service repeatedly restarts or fails to start.

**Diagnosis**:

```bash
# Check service logs
docker compose -f deploy/docker-compose/docker-compose.dev.yml logs <service-name>

# Verify build locally
go build ./apps/<service-name>/...
```

### Port Conflicts

**Symptom**: Startup reports that ports are already in use.

**Diagnosis**:

```bash
# Check port usage
lsof -i :8080
lsof -i :3000
```

**Resolution**: Stop the process occupying the port, or update the port configuration in `.env`.

## Database Issues

### Migration Failures

**Symptom**: Service reports missing tables or fields on startup.

**Resolution**:

```bash
# Re-run migrations
make dev-down
make migrate
make dev-up
```

### Connection Refused

**Symptom**: Service logs show `connection refused`.

**Diagnosis**:

```bash
# Confirm PostgreSQL container is running
docker ps | grep postgres

# Check database readiness
docker compose -f deploy/docker-compose/docker-compose.dev.yml exec postgres pg_isready
```

## Frontend Issues

### Blank Page / Failed to Load

**Diagnosis**:

1. Confirm Vite dev server is running: `docker ps | grep web`
2. Check browser console for network errors
3. Confirm gateway is reachable: `curl http://localhost:8080/api/v1/projects`

### Language Switching Not Working

**Diagnosis**:

1. Check if browser localStorage has been cleared
2. Verify that locale state is correct in `localeStore`
3. Verify that the relevant key exists in `messages.ts`

## Agent Issues

### Agent Cannot Claim Tasks

**Diagnosis**:

```bash
# Check agent logs
docker compose -f deploy/docker-compose/docker-compose.dev.yml logs agent

# Verify Redis queue
docker compose -f deploy/docker-compose/docker-compose.dev.yml exec redis redis-cli PING
```

### Docker Image Pull Failures

**Diagnosis**:

1. Confirm image is built: `docker images | grep crawler`
2. Check if the agent container can access the Docker socket
3. For private registries, verify `IMAGE_REGISTRY_AUTH_MAP` configuration

## Test Issues

### Test Timeouts

**Symptom**: Some tests timeout during `make test`.

**Resolution**:

```bash
# Run the failing test package individually with a longer timeout
go test -timeout 60s ./apps/<service-name>/...
```

### Frontend Test Failures

**Diagnosis**:

```bash
# Run frontend tests inside the container
npm --prefix apps/web test
```
