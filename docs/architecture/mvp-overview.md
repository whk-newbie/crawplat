# MVP Architecture Overview

## Current State

The MVP foundation is implemented as a small control plane plus execution-plane agent:

- `gateway` provides the single public API entry point.
- `iam-service` handles login and JWT issuance.
- `project-service` owns project creation and listing.
- `spider-service` owns project-scoped spider creation and listing.
- `execution-service` owns manual execution records and execution logs.
- `node-service` tracks node heartbeats and current node inventory.
- `datasource-service` owns datasource configuration, connection checks, and previews.
- `agent` runs on a node and sends heartbeats.

The stack is designed for Docker Compose first. PostgreSQL, Redis, and MongoDB are started together with the API services and the web shell so the MVP environment matches the planned platform shape. The current service implementations still use in-memory stores for the vertical slice, so the datastores are provisioned but not yet wired into the request path.

## Runtime Shape

The gateway exposes the `/api/v1/*` surface to the outside world and proxies requests to the matching service. Services own their own handlers and store their own domain state, which keeps the implementation testable and makes the boundaries visible even in the MVP.

The agent talks directly to the node service for heartbeat updates. That keeps node liveness independent from user-facing API traffic and gives the execution plane a separate lifecycle from the control plane.

## What Is Implemented

- login via `POST /api/v1/auth/login`
- project create/list slice
- project-scoped spider create/list slice
- execution lifecycle slice for create, fetch, and logs
- node heartbeat tracking and node listing
- datasource create, list, test, and preview
- gateway routing for the current API surface
- Compose wiring for the services, datastores, and web shell

## What This MVP Does Not Yet Cover

- scheduled execution orchestration
- task DAGs or workflow orchestration
- multi-tenant isolation
- advanced RBAC beyond the current login slice
- production hardening beyond the Compose-based MVP stack
