# Phase 3 Smoke Checklist

## Goal

Verify the Phase 3 additions on top of the working Phase 2 execution pipeline:

- `scheduler-service` is part of the Compose stack
- schedules can be created and listed through the gateway
- scheduled executions carry retry configuration into execution records
- retry materialization endpoint can create a retry execution from a failed execution
- monitor overview route and web page are reachable

## Preconditions

- Docker is available
- sample images already exist:
  - `crawler/go-echo:latest`
  - `crawler/python-echo:latest`

## Bring Up

```bash
make migrate
make up
```

## Schedule API

Create a project and spider first, then create a schedule through the gateway:

```bash
curl -sS -X POST http://localhost:8080/api/v1/schedules \
  -H 'Content-Type: application/json' \
  -d '{
    "projectId":"<project-id>",
    "spiderId":"<spider-id>",
    "name":"nightly-go-echo",
    "cronExpr":"*/5 * * * *",
    "enabled":true,
    "image":"crawler/go-echo:latest",
    "command":["./go-echo"],
    "retryLimit":2,
    "retryDelaySeconds":30
  }'
```

Expected:

- `201 Created`
- response includes `retryLimit: 2`
- response includes `retryDelaySeconds: 30`

List schedules:

```bash
curl -sS http://localhost:8080/api/v1/schedules
```

Expected:

- schedule appears in the list

## Retry Materialization

This endpoint is internal and requires `X-Internal-Token`.

```bash
curl -i -X POST http://localhost:8085/internal/v1/executions/retries/materialize \
  -H 'X-Internal-Token: change-me'
```

Expected:

- `201 Created` when a failed eligible execution exists
- `204 No Content` when no failed eligible execution exists

If `201 Created`, confirm the returned execution has:

- `triggerSource: "retry"`
- incremented `retryCount`
- `retryOfExecutionId` pointing at the failed execution

## Monitor Route

```bash
curl -sS http://localhost:8080/api/v1/monitor/overview
```

Expected:

- `200 OK`
- JSON shape contains `executions` and `nodes`
- execution counters reflect rows currently stored in `executions`
- node counters reflect the current Redis heartbeat set

Note:

- offline node count remains `0` in the MVP because only live heartbeats are persisted

## Web Route

Open:

- `http://localhost:3000/monitor`

Expected:

- page loads
- refresh button works
- raw payload section renders the monitor overview response

## Tear Down

```bash
make down
```
