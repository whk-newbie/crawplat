# Local Development Guide

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker and Docker Compose
- Make

## Quick Start

### 1. Clone the Repository

```bash
git clone <repo-url> crawler-platform
cd crawler-platform
```

### 2. Start the Development Environment

```bash
make dev-up
```

This command:
- Starts PostgreSQL and runs database migrations
- Starts all services in development mode (Go services hot-reload via `air`)
- Starts the frontend Vite dev server

### 3. Access the Application

- Frontend: `http://localhost:3000`
- Gateway API: `http://localhost:8080`

### 4. Stop the Environment

```bash
make dev-down
```

## Project Structure

```
.
├── apps/                   # Application services
│   ├── gateway/            # API Gateway
│   ├── iam-service/        # Authentication service
│   ├── project-service/    # Project service
│   ├── spider-service/     # Spider service
│   ├── execution-service/  # Execution service
│   ├── scheduler-service/  # Schedule service
│   ├── node-service/       # Node service
│   ├── datasource-service/ # Datasource service
│   ├── monitor-service/    # Monitor service
│   └── web/                # Frontend application
├── deploy/                 # Deployment configuration
│   ├── docker-compose/     # Compose files
│   ├── migrations/         # Database migrations
│   └── scripts/            # Operations scripts
├── docs/                   # Documentation
├── examples/               # Example spiders
│   └── spiders/
└── packages/               # Shared Go packages
    └── go-common/
```

## Development Workflow

### Backend Development

When you modify Go service code, `air` automatically recompiles and restarts the corresponding service. Changes to the following paths trigger reloads:

- `apps/*` (all services)
- `packages/go-common` (shared packages)

### Frontend Development

The frontend uses the Vite dev server with Hot Module Replacement (HMR). Changes to files under `apps/web/src` trigger automatic browser refresh.

### API Proxying

The frontend dev server proxies `/api/*` requests to the `gateway` container in the Docker network. No manual API address configuration is needed.

### Database Migrations

If you modify database schemas, add migration files:

```bash
# Migration files are located at
deploy/migrations/postgres/
```

Then restart the development environment:

```bash
make dev-down && make dev-up
```

## Running Tests

### All Tests

```bash
make test
```

### Backend Tests Only

```bash
go test ./...
```

### Frontend Tests Only

```bash
npm --prefix apps/web test
```

## Building Example Spider Images

```bash
docker build -t crawler/go-echo:latest examples/spiders/go-echo
docker build -t crawler/python-echo:latest examples/spiders/python-echo
```

## Full Build and Smoke Verification

To validate the release configuration, use:

```bash
make migrate
make up
bash deploy/scripts/smoke-mvp.sh
```

See `docs/product/mvp-smoke-checklist.en-US.md` for details.
