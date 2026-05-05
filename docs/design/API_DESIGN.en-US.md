# API Design Document (2026-05 Snapshot)

## Overview

The Crawler Platform API is divided into two categories:

- Public API: `/api/v1/*`, exposed uniformly by the gateway to the Web frontend and external clients.
- Internal API: `/internal/v1/executions/*`, used by agent / scheduler to interact with execution-service.

All endpoints use JSON.

## Authentication and Gateway Strategy

- JWT: The gateway implements Bearer Token validation (`GATEWAY_ENFORCE_JWT` configurable).
  - Only protects public business APIs; does not intercept `/api/<version>/auth/*`.
  - On failure, returns `{ "error": "message" }` uniformly, with the following scenarios distinguished:
    - Missing Bearer Token: `missing bearer token`
    - Invalid Authorization scheme: `invalid authorization scheme`
    - Invalid Bearer Token: `invalid bearer token`
    - `JWT_SECRET` not configured but JWT enforcement enabled: `jwt secret is not configured`
- Internal Token: `X-Internal-Token`, value from `INTERNAL_API_TOKEN` (falls back to `JWT_SECRET` when unset).
  - Used to protect `/internal/v1/executions/*`.
  - On failure, returns:
    - Token not configured: `internal token is not configured`
    - Token incorrect: `unauthorized internal route`
- Gateway enhancements:
  - Rate limiting (`GATEWAY_RATE_LIMIT_*`)
  - Request-ID and access logging
    - Default request-id header is `X-Request-ID`, overridable via `GATEWAY_REQUEST_ID_HEADER`.
    - When `GATEWAY_TRUST_REQUEST_ID=true`, the gateway preferentially forwards the client-provided request-id; otherwise it generates one.
    - The gateway writes request-id into request headers, response headers, and the forwarding chain.
  - API version routing (default `v1`, extensible via `GATEWAY_API_SUPPORTED_VERSIONS`)
    - Configured and supported versions are exposed at the gateway layer and can route to the current upstream implementation by stable version.
    - Unsupported versions or unknown API routes return `404`: `unsupported api version or route`

## Common Response Convention

Errors are uniformly:

```json
{ "error": "message" }
```

Common status codes: `200`, `201`, `204`, `400`, `401`, `404`, `409`, `500`, `502`.

## Public API Listing

### iam-service

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### project-service

- `POST /api/v1/projects`
- `GET /api/v1/projects?limit=<n>&offset=<n>`

> The gateway currently only explicitly registers `/api/<version>/projects`. Project detail sub-paths have not yet been expanded in this gateway worktree.

### spider-service

- `POST /api/v1/projects/:projectId/spiders`
- `GET /api/v1/projects/:projectId/spiders?limit=<n>&offset=<n>`
- `GET /api/v1/projects/:projectId/registry-auth-refs`
- `POST /api/v1/spiders/:spiderId/versions`
- `GET /api/v1/spiders/:spiderId/versions`

### execution-service

- `POST /api/v1/executions`
- `GET /api/v1/executions?limit=<n>&offset=<n>&status=<string>`
- `GET /api/v1/executions/:id`
- `GET /api/v1/executions/:id/logs`
- `POST /api/v1/executions/:id/logs`

`GET /api/v1/executions` supports pagination via `limit` (default 20, max 100) and `offset` (default 0), and optional `status` filtering.

`POST /api/v1/executions` supports:

- `spiderVersion`
- `registryAuthRef`
- Resource limits: `cpuCores` / `memoryMB` / `timeoutSeconds`
- Retry fields: `retryLimit` / `retryCount` / `retryDelaySeconds` / `retryOfExecutionId`

When a request does not explicitly provide `registryAuthRef`, execution-service attempts to inherit it from the specified spider version.

### scheduler-service

- `POST /api/v1/schedules`
- `GET /api/v1/schedules?limit=<n>&offset=<n>`

Scheduled tasks support: `spiderVersion`, `registryAuthRef`, retry strategy fields.

### node-service

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes?limit=<n>&offset=<n>`
- `GET /api/v1/nodes/:id`
- `GET /api/v1/nodes/:id/sessions?limit=<n>&gapSeconds=<n>`

`GET /api/v1/nodes/:id` supports execution history filter parameters:

- `executionLimit` / `executionOffset`
- `executionStatus`
- `executionFrom` / `executionTo` (RFC3339)

### datasource-service

- `POST /api/v1/datasources`
- `GET /api/v1/datasources?projectId=<id>&limit=<n>&offset=<n>`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`

`test/preview` are now real probes, no longer fixed mock responses.

### monitor-service

- `GET /api/v1/monitor/overview`
- `POST /api/v1/monitor/alerts/rules`
- `GET /api/v1/monitor/alerts/rules`
- `PATCH /api/v1/monitor/alerts/rules/:id`
- `GET /api/v1/monitor/alerts/events?limit=<n>&offset=<n>`

Alert rules support two types: execution failure and node offline. Alert events are persisted to PostgreSQL.

## Internal API (execution-service)

The following endpoints are dual-validated by gateway and execution-service with the internal token:

- `POST /internal/v1/executions/claim`
- `POST /internal/v1/executions/:id/start`
- `POST /internal/v1/executions/:id/logs`
- `POST /internal/v1/executions/:id/complete`
- `POST /internal/v1/executions/:id/fail`
- `POST /internal/v1/executions/retries/materialize`

## Private Registry Integration

- `registryAuthRef` can be propagated across spider versions / schedules / executions.
- On startup, the agent reads `IMAGE_REGISTRY_AUTH_MAP`:
  - Supports host-key: `"<registry-host>" -> credential`
  - Supports named-ref: `"<registryAuthRef>" -> { server: "<registry-host>", ... }`
- When the image's registry host matches, or the execution's `registryAuthRef` hits a mapping, the agent performs `docker login` + `docker pull`.
