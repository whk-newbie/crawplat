# MVP Progress Update

Date: 2026-04-12
Branch: `feat/mvp-foundation`
Worktree: `/home/iambaby/goland_projects/crawler-platform/.worktrees/mvp-foundation`

## Summary

The MVP foundation implementation is complete at the code and documentation level. All twelve planned tasks are now implemented and reviewed:

- repository skeleton
- shared Go foundation package
- IAM login slice
- project-service CRUD slice
- spider-service create/list slice
- node-service and agent heartbeat loop
- execution-service manual execution and log ingest slice
- datasource-service CRUD, connection test, and preview slice
- gateway routing and auth passthrough
- Vue web MVP shell
- Docker Compose stack wiring and smoke flow
- onboarding and architecture docs

## Completed Components

### Repository Foundation

- root `Makefile`
- `go.work`
- root `go.mod`
- `.env.example`
- initial Docker Compose skeleton
- dynamic nested-module test execution from root `make test`

### Shared Go Package

- JWT issue/parse helpers
- config struct
- thin HTTP response helper
- thin PostgreSQL / Redis / Mongo wrappers

### IAM Service

- `POST /api/v1/auth/login`
- seeded admin path gated behind explicit env opt-in
- JWT secret required at startup
- HTTP and service-level tests for success and negative paths

### Project Service

- `POST /api/v1/projects`
- `GET /api/v1/projects`
- in-memory project store
- JSON response shape tests
- service and router coverage

### Spider Service

- `POST /api/v1/projects/:projectId/spiders`
- `GET /api/v1/projects/:projectId/spiders`
- language/runtime validation
- project-scoped list behavior
- service and router coverage

### Node Service and Agent

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes`
- in-memory node heartbeat tracking
- agent startup heartbeat
- heartbeat failure surfacing
- graceful signal-driven shutdown
- router and agent heartbeat tests

### Execution Service

- `POST /api/v1/executions`
- `POST /api/v1/executions/:id/logs`
- `GET /api/v1/executions/:id`
- `GET /api/v1/executions/:id/logs`
- manual execution record bootstrap with `pending` state
- in-memory execution and log storage
- service and router happy-path coverage

### Datasource Service

- `POST /api/v1/datasources`
- `GET /api/v1/datasources`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`
- datasource type validation for `mongodb`, `redis`, and `postgresql`
- read-only preview behavior with bounded in-memory responses
- service and router coverage

### Gateway

- `ANY /api/v1/auth`
- `ANY /api/v1/auth/*path`
- `ANY /api/v1/projects`
- `ANY /api/v1/projects/:projectId/spiders`
- `ANY /api/v1/executions`
- `ANY /api/v1/executions/*path`
- `ANY /api/v1/nodes`
- `ANY /api/v1/nodes/*path`
- `ANY /api/v1/datasources`
- `ANY /api/v1/datasources/*path`
- JWT passthrough for protected routes
- route-level proxy coverage

### Web MVP Shell

- Vue 3 + TypeScript + Vite shell
- login page and token persistence
- dashboard summary
- projects view
- spiders view
- executions view
- datasources view
- Vitest coverage and inclusion in root `make test`

### Compose and Smoke Flow

- service Dockerfiles for the Go services and web shell
- `deploy/docker-compose/docker-compose.mcp.yml` with datastores and service containers
- `make up` build-and-start workflow
- `make down` teardown workflow
- `deploy/scripts/smoke-mvp.sh`
- `docs/product/mvp-smoke-checklist.md`

### Onboarding and Architecture Docs

- root `README.md`
- `docs/architecture/mvp-overview.md`
- `docs/api/mvp-service-map.md`

## Current Notes

- The codebase is being developed in an isolated git worktree and has not yet been merged back to the project root branch.
- The current stopping point is clean at the repository level after Task 12.
- `make test` passes for the implemented MVP slice.

## Phase 2 Addendum

As of 2026-04-22, the MVP foundation has advanced beyond the original 2026-04-12 snapshot:

- `project-service`, `spider-service`, and `datasource-service` now persist core metadata in PostgreSQL.
- `node-service` stores heartbeat liveness in Redis.
- `execution-service` stores execution metadata in PostgreSQL, execution logs in MongoDB, and queue state in Redis.
- internal execution lifecycle routes now cover claim, start, append-log, complete, and fail flows behind `X-Internal-Token`.
- `make test` still passes after the persistence and execution-lifecycle work.
- Full `make up` verification is blocked in this environment because Docker bridge networking is not supported here (`failed to add the host <=> sandbox pair interfaces: operation not supported`).
