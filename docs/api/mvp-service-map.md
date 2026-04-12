# MVP Service Map

## Service Responsibilities

- `gateway` -> public entry and request proxying
- `iam-service` -> auth
- `project-service` -> project CRUD
- `spider-service` -> spider CRUD
- `execution-service` -> manual execution and logs
- `node-service` -> node heartbeat
- `datasource-service` -> datasource config and preview
- `agent` -> heartbeat and execution pull

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

### `node-service`

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes`

### `datasource-service`

- `POST /api/v1/datasources`
- `GET /api/v1/datasources`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`

## Notes

- The gateway is the only public API surface expected by the web app and external callers.
- The agent currently uses the node heartbeat route directly so node liveness stays independent of gateway traffic.
