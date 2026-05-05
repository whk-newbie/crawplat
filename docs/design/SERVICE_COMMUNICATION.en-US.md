# Service Communication Design (2026-05 Snapshot)

## Overall Topology

- Browser -> web (Vite/Nginx)
- Browser -> gateway (`/api/v1/*`)
- gateway -> business services (reverse proxy)
- agent connects directly to node-service and execution-service internal endpoints
- scheduler-service calls execution-service public/internal endpoints
- monitor-service aggregates PostgreSQL + Redis

## Communication and Authentication Matrix

| Caller | Callee | Path | Authentication |
|---|---|---|---|
| Browser | Gateway | `/api/v1/*` | Bearer JWT (configurable enforcement) |
| Gateway | iam-service | `/api/v1/auth/*` | Pass-through |
| Gateway | project-service | `/api/v1/projects*` | Pass-through |
| Gateway | spider-service | `/api/v1/projects/:id/spiders*`, `/api/v1/spiders/:id/versions*` | Pass-through |
| Gateway | execution-service | `/api/v1/executions*` | Pass-through |
| Gateway | execution-service | `/internal/v1/executions*` | `X-Internal-Token` |
| Gateway | node-service | `/api/v1/nodes*` | Pass-through |
| Gateway | datasource-service | `/api/v1/datasources*` | Pass-through |
| Gateway | scheduler-service | `/api/v1/schedules*` | Pass-through |
| Gateway | monitor-service | `/api/v1/monitor*` | Pass-through |
| Agent | node-service | `/api/v1/nodes/:id/heartbeat` | None |
| Agent | execution-service | `/internal/v1/executions/*` | `X-Internal-Token` |
| Scheduler | execution-service | `/api/v1/executions` | None (internal network call) |
| Scheduler | execution-service | `/internal/v1/executions/retries/materialize` | `X-Internal-Token` |

## Execution Flow

1. User calls `POST /api/v1/executions` to create an execution (writes PG + enqueues to Redis pending).
2. Agent polls `POST /internal/v1/executions/claim` to claim a task.
3. Agent calls `start`, runs the container and appends logs to `/logs`.
4. After execution completes, agent calls `complete` or `fail`.
5. Scheduler periodically calls `retries/materialize` to materialize retry tasks.

## Scheduling Flow

1. Scheduler scans enabled tasks and evaluates cron expressions.
2. Uses `last_materialized_at` cursor to advance and prevent duplicates.
3. On reaching a trigger window, calls the execution create API to generate an execution.

## Alerting Flow

1. monitor-service polls for failed executions and offline nodes.
2. Matches against `alert_rules` by rule type.
3. Sends webhook; results are persisted to `alert_events`.
4. Frontend queries historical events via `/api/v1/monitor/alerts/events` with pagination.

## Private Image Pull Flow

1. Execution records may carry `registryAuthRef` (or inherit from spider version).
2. Agent reads registry credential mappings from `IMAGE_REGISTRY_AUTH_MAP`.
3. When the image host matches a mapping entry, agent performs `docker login`/`docker pull` before `docker run`.
