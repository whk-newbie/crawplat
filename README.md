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
- `make up` builds the Linux service binaries, builds the web assets, runs `make migrate`, and starts the full Compose stack with PostgreSQL, Redis, MongoDB, gateway, iam-service, project-service, spider-service, execution-service, node-service, datasource-service, agent, and web.
- On a normal Docker host, the intended smoke flow is: login, create a project, register a Docker spider, create a manual execution, and inspect execution detail plus logs in the web app.
- This flow was exercised on 2026-04-22 with both `crawler/go-echo:latest` and `crawler/python-echo:latest`; both executions reached `succeeded` and returned logs through `GET /api/v1/executions/:id/logs`.
