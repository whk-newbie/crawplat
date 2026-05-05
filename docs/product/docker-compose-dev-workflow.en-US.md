# Docker Compose Development Workflow

This project maintains two Compose workflows:

- `deploy/docker-compose/docker-compose.mcp.yml`
  A release-oriented runtime configuration, suitable for full builds and verification.
- `deploy/docker-compose/docker-compose.dev.yml`
  A development-oriented runtime configuration, suitable for live editing and debugging.

## Use Cases

For daily development, prefer the development workflow. It addresses two problems:

- Go services automatically recompile and restart after code changes
- Frontend changes take effect immediately through Vite hot reload

## Starting

Run from the repository root:

```bash
make dev-up
```

This command first starts PostgreSQL in the dev stack, runs migrations, then brings up the full development environment.

## Stopping

```bash
make dev-down
```

## Ports

- `http://localhost:8080`
  External gateway entry point
- `http://localhost:3000`
  Frontend Vite dev server

Open `http://localhost:3000` in a browser.

## Hot Reload

### Go Services

Go services run inside a unified development image using `air` to watch the source tree. Changes to these paths trigger automatic recompilation and restart:

- `apps/*`
- `packages/go-common`

### Frontend

The frontend uses the Vite dev server. Changes to files under `apps/web/src` trigger automatic page refresh or hot module replacement.

## API Proxying

The frontend dev server proxies `/api/*` requests to the `gateway` container inside the Docker network, so frontend code does not need hardcoded container addresses.

That means:

- The browser accesses `localhost:3000`
- Vite (inside the container) forwards `/api/*` to `http://gateway:8080`

## Development Tips

- For daily frontend-backend integration, prefer `make dev-up`
- When you need to validate the release configuration, use `make up`
- If you change database schemas, add migration files first, then re-run `make dev-up`

## Current Limitations

- This development workflow prioritizes local development efficiency, not production deployment
- The `agent` still depends on the host Docker socket
- This workflow is not wired into remote CI — it is a local containerized development flow
