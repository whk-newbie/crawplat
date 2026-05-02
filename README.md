# Crawler Platform

## Developer Checklist

- install Go and Node.js
- inspect `deploy/env/.env.example` for the intended runtime variables
- build the sample execution images:
  - `docker build -t crawler/go-echo:latest examples/spiders/go-echo`
  - `docker build -t crawler/python-echo:latest examples/spiders/python-echo`
- run `make migrate`
- run `make up`
- run `make test`

## Developer Workflow

- `make test` runs the root Go test suite, then the nested Go module test suites under `packages/` and `apps/`, and finally `apps/web` with `npm test`.
- `make migrate` starts the PostgreSQL container, waits for readiness, and applies the SQL migrations under `deploy/migrations/postgres`.
- `make up` builds the Linux service binaries, builds the web assets, runs `make migrate`, and starts the full Compose stack with PostgreSQL, Redis, MongoDB, gateway, iam-service, project-service, spider-service, execution-service, node-service, datasource-service, scheduler-service, monitor-service, agent, and web.
- `make dev-up` starts a dedicated development Compose workflow with PostgreSQL, Redis, MongoDB, every Go service running through containerized hot reload, and the web app served by Vite on port `3000`.
- `make dev-down` tears down that development workflow.
- On a normal Docker host, the intended smoke flow is: login, create a project, register a Docker spider, create a manual execution, and inspect execution detail plus logs in the web app.
- This flow was exercised on 2026-04-22 with both `crawler/go-echo:latest` and `crawler/python-echo:latest`; both executions reached `succeeded` and returned logs through `GET /api/v1/executions/:id/logs`.
- Phase 3 adds scheduled execution materialization, retry-policy orchestration, and a monitor overview route plus web page. The monitor overview is now backed by aggregated execution status counts from PostgreSQL and live node counts from Redis. Offline nodes still report as `0` because the MVP node inventory only persists currently live heartbeats.

## Current Phase 7 Progress

- Datasource `test` and `preview` now run real probes instead of fixed mock responses.
- Monitor alerting now supports:
  - rule create/list/update APIs
  - webhook delivery event persistence
  - split polling cadence (`MONITOR_ALERT_POLL_INTERVAL` and `MONITOR_NODE_OFFLINE_ALERT_POLL_INTERVAL`) so node-offline alerts can be checked more frequently.
- Spider version management is now available:
  - `POST /api/v1/spiders/:spiderId/versions`
  - `GET /api/v1/spiders/:spiderId/versions`
  - spider versions now persist optional `registryAuthRef` for private registry credentials.
  - `SpidersView` supports listing versions, setting version-level `registryAuthRef`, and creating a new version.
- Private image pull integration is now available in agent runtime:
  - set `IMAGE_REGISTRY_AUTH_MAP` as JSON map:
    - host-key mode: `{ "<registry-host>": { "username": "...", "password": "...", "server": "..." } }`
    - named-ref mode (for `registryAuthRef` such as `ghcr-prod`): `{ "<auth-ref>": { "server": "<registry-host>", "username": "...", "password": "..." } }`
  - when execution image host matches configured registry, or execution carries a matched `registryAuthRef`, agent runs `docker login` + `docker pull` before `docker run`.
  - execution/schedule APIs now support optional `registryAuthRef` to select a named credential directly.
  - execution creation now auto-inherits `registryAuthRef` from resolved spider version when request-level `registryAuthRef` is empty.
  - spider-service now exposes `GET /api/v1/projects/:projectId/registry-auth-refs` for project-scoped saved refs; web forms can load these refs to reduce manual typing.

## Containerized Dev Workflow

- `make dev-up` uses `deploy/docker-compose/docker-compose.dev.yml`
- Go services run inside a shared dev image with `air`, so editing `apps/*` or `packages/go-common` triggers an in-container rebuild and restart
- the web app runs with `vite` on `http://localhost:3000`
- `/api/*` requests from Vite are proxied to the gateway container on `http://gateway:8080`
- Gateway upstreams resolve to Compose DNS by default; override any service with `GATEWAY_UPSTREAM_<SERVICE>`.
- Chinese guide: `docs/product/docker-compose-dev-workflow.zh-CN.md`
