# Frontend Design Document (2026-05 Snapshot)

## Tech Stack

- Vue 3 + TypeScript
- Vue Router
- Pinia
- Element Plus
- Vite
- Vitest

## Pages and Routes

- `/login` Login
- `/projects` Project management
- `/spiders` Spider and version management
- `/executions` Manual execution creation
- `/executions/:id` Execution detail and logs
- `/schedules` Schedule management
- `/datasources` Datasource management
- `/monitor` Monitoring and alerting
- `/nodes` Node inventory and history

## Current Implementation Status

All of the above pages have been implemented. There are no remaining placeholder pages.

## Key Interactions

### Spiders

- Supports creating spiders
- Supports version listing and version creation
- Versions can set `registryAuthRef`
- Supports "Load Refs" to pull project-level registry auth refs

### Executions

- Supports creating executions by project/spider/version
- Supports `registryAuthRef` and resource limit fields
- `registryAuthRef` input supports an optional dropdown (can create new values)

### Schedules

- Supports cron, enable/disable, and retry strategy
- Supports `spiderVersion` + `registryAuthRef`
- `registryAuthRef` also supports loading project refs

### Monitor

- Overview statistics cards
- Alert rule creation / listing / updating
- Alert event paginated listing

### Nodes

- Paginated node listing
- Node detail (recent heartbeats, execution history filtering)
- Sessions aggregation view

### Datasources

- Datasource creation and listing
- Real test/preview probe result display

## API Layer

Directory: `apps/web/src/api/*`

- All requests are uniformly wrapped by `client.ts`
- Default to gateway `/api/v1/*`
- List endpoints uniformly return paginated structures of `{ items, total, limit, offset }` (where backend supports it)

## Testing

- View-level tests located in `apps/web/src/views/__tests__`
- Currently covers key pages: projects, spiders, executions, schedules, datasources, monitor, nodes, login
