# Crawler Platform

## Developer Checklist

- install Go and Node.js
- copy `deploy/env/.env.example`
- run `make up`
- run `make test`

## Developer Workflow

- `make test` runs the root Go test suite, then the nested Go module test suites under `packages/` and `apps/`, and finally `apps/web` with `npm test`.
- `make up` builds the Linux service binaries, builds the web assets, and starts the full Compose stack with PostgreSQL, Redis, MongoDB, gateway, iam-service, project-service, spider-service, execution-service, node-service, datasource-service, agent, and web.
- On a normal Docker host, `make up` is expected to bring the MVP stack up cleanly and the services should report healthy.
- In this worktree, full Compose verification is currently blocked by the sandbox Docker bridge networking limitation (`failed to add the host <=> sandbox pair interfaces: operation not supported`).
