# MVP Progress Update

Date: 2026-04-12
Branch: `feat/mvp-foundation`
Worktree: `/home/iambaby/goland_projects/crawler-platform/.worktrees/mvp-foundation`

## Summary

The MVP foundation is in progress. Core repository scaffolding and the first six implementation tasks are complete and reviewed:

- repository skeleton
- shared Go foundation package
- IAM login slice
- project-service CRUD slice
- spider-service create/list slice
- node-service and agent heartbeat loop

This leaves the execution path, datasource slice, gateway, frontend shell, stack wiring, and final onboarding docs still to be completed.

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

## Remaining Tasks

- Task 7: execution-service manual execution and log ingest
- Task 8: datasource-service MVP
- Task 9: gateway routing
- Task 10: Vue web shell
- Task 11: Docker Compose full wiring and smoke flow
- Task 12: onboarding and architecture docs

## Current Notes

- The codebase is being developed in an isolated git worktree and has not yet been merged back to the project root branch.
- The current stopping point is clean: Task 6 is complete and reviewed with no open Important/Critical issues.
- The next implementation step is Task 7, which will build the first execution record lifecycle on top of the completed node heartbeat foundation.
