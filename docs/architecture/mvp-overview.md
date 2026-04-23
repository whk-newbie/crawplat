# MVP Architecture Overview

## Current State

The MVP foundation is implemented as a small control plane plus execution-plane agent:

- `gateway` provides the single public API entry point.
- `iam-service` handles login and JWT issuance.
- `project-service` owns project creation and listing backed by PostgreSQL.
- `spider-service` owns project-scoped spider creation and listing backed by PostgreSQL.
- `execution-service` owns manual execution records in PostgreSQL, logs in MongoDB, and execution queue coordination through Redis.
- `execution-service` also owns scheduled and retry execution materialization metadata.
- `node-service` tracks node heartbeats and current node inventory through Redis.
- `datasource-service` owns datasource configuration, connection checks, and previews backed by PostgreSQL.
- `scheduler-service` owns persisted schedules in PostgreSQL, cron polling, scheduled execution creation, and retry orchestration triggers.
- `monitor-service` exposes the first monitor overview route. It is currently a scaffold service with the final route shape but not yet backed by aggregated persistence queries.
- `agent` runs on a node, sends heartbeats, polls executions, and reports lifecycle/log updates back to the control plane.

The stack is designed for Docker Compose first. PostgreSQL, Redis, and MongoDB are started together with the API services and the web shell so the MVP environment matches the planned platform shape. The current slice already uses those datastores on the request path for projects, spiders, datasources, executions, logs, queue state, and node liveness.

## Runtime Shape

The gateway exposes the `/api/v1/*` surface to the outside world and proxies requests to the matching service. Services own their own handlers and store their own domain state, which keeps the implementation testable and makes the boundaries visible even in the MVP.

The gateway exposes the public `/api/v1/*` surface and selected internal `/internal/v1/executions/*` routes for execution workers and scheduler-triggered retry materialization. Internal execution routes are protected with `X-Internal-Token`, while user-facing routes continue through the public API surface.

The agent talks directly to the node service for heartbeat updates. That keeps node liveness independent from user-facing API traffic and gives the execution plane a separate lifecycle from the control plane. The execution worker flow now supports queue claim, start, append-log, complete, and fail transitions, and the agent is wired to consume that path through a Docker-based runner.

## What Is Implemented

- login via `POST /api/v1/auth/login`
- project create/list slice
- project-scoped spider create/list slice
- execution lifecycle slice for create, fetch, logs, claim, start, complete, and fail
- scheduled execution materialization and retry-policy orchestration
- node heartbeat tracking and node listing backed by Redis
- datasource create, list, test, and preview
- monitor overview route and web monitor page shell
- gateway routing for the current API surface
- Compose wiring for the services, datastores, web shell, migration flow, and Docker-capable agent runtime
- verified Go/Python Docker execution path through the agent poller

## What This MVP Does Not Yet Cover

- task DAGs or workflow orchestration
- multi-tenant isolation
- advanced RBAC beyond the current login slice
- production hardening beyond the Compose-based MVP stack
- monitor aggregation backed by execution/node summary queries instead of the current scaffold response
