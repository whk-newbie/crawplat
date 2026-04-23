# MVP Service Map

## Service Responsibilities

- `gateway` -> public entry and request proxying
- `iam-service` -> auth
- `project-service` -> project create/list in PostgreSQL
- `spider-service` -> spider create/list in PostgreSQL
- `execution-service` -> execution metadata in PostgreSQL, logs in MongoDB, and queue-backed lifecycle transitions
- `execution-service` -> execution metadata in PostgreSQL, logs in MongoDB, queue-backed lifecycle transitions, and retry materialization
- `node-service` -> node heartbeat in Redis
- `datasource-service` -> datasource config and preview
- `scheduler-service` -> schedule persistence, cron polling, scheduled execution creation, retry trigger loop
- `monitor-service` -> monitor overview route scaffold
- `agent` -> heartbeat, execution polling, Docker runtime execution

## Key Routes

### `gateway`

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
- `ANY /api/v1/schedules`
- `ANY /api/v1/monitor/*path`
- `ANY /internal/v1/executions/claim`
- `ANY /internal/v1/executions/:id/start`
- `ANY /internal/v1/executions/:id/logs`
- `ANY /internal/v1/executions/:id/complete`
- `ANY /internal/v1/executions/:id/fail`
- `ANY /internal/v1/executions/retries/materialize`

### `iam-service`

- `POST /api/v1/auth/login`

### `project-service`

- `POST /api/v1/projects`
- `GET /api/v1/projects`

### `spider-service`

- `POST /api/v1/projects/:projectId/spiders`
- `GET /api/v1/projects/:projectId/spiders`

### `execution-service`

- `POST /api/v1/executions`
- `POST /api/v1/executions/:id/logs`
- `GET /api/v1/executions/:id`
- `GET /api/v1/executions/:id/logs`
- `POST /internal/v1/executions/claim`
- `POST /internal/v1/executions/:id/start`
- `POST /internal/v1/executions/:id/logs`
- `POST /internal/v1/executions/:id/complete`
- `POST /internal/v1/executions/:id/fail`
- `POST /internal/v1/executions/retries/materialize`

### `node-service`

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes`

### `datasource-service`

- `POST /api/v1/datasources`
- `GET /api/v1/datasources`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`

### `scheduler-service`

- `POST /api/v1/schedules`
- `GET /api/v1/schedules`

### `monitor-service`

- `GET /api/v1/monitor/overview`

## Notes

- The gateway is the only public API surface expected by the web app and external callers.
- Internal execution routes are reserved for execution workers and require `X-Internal-Token`.
- The agent uses the node heartbeat route directly for liveness and the execution-service internal routes for claim/start/log/complete/fail flow.
- The scheduler service uses the public execution create route for scheduled executions and the internal retry materialization route for retry orchestration.
- The monitor service route is wired end-to-end, but its response is still scaffold-level and not yet backed by aggregate persistence queries.
