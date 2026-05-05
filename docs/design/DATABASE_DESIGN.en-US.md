# Database Design Document (2026-05 Snapshot)

## Overview

The platform uses three storage systems:

- PostgreSQL: Business master data and alert events
- MongoDB: Execution logs
- Redis: Execution queue and online node transient state

## PostgreSQL

Migration directory: `deploy/migrations/postgres`

Current migrations (in filename order):

- `001_phase2_core_tables.sql`
- `002_phase2_execution_indexes.sql`
- `003_phase3_scheduled_tasks.sql`
- `004_phase3_schedule_materialization.sql`
- `005_phase3_retry_policies.sql`
- `006_phase4_nodes.sql`
- `007_phase7_execution_resource_limits.sql`
- `007_phase7_iam_users.sql`
- `008_phase7_alerting.sql`
- `009_phase7_spider_versions.sql`
- `010_phase7_execution_spider_version.sql`
- `011_phase7_scheduled_task_spider_version.sql`
- `012_phase7_execution_registry_auth_ref.sql`
- `013_phase7_scheduled_task_registry_auth_ref.sql`
- `014_phase7_spider_version_registry_auth_ref.sql`

### Core Tables

- `projects`
- `spiders`
- `spider_versions`
- `datasources`
- `executions`
- `scheduled_tasks`
- `nodes`
- `node_heartbeats`
- `alert_rules`
- `alert_events`
- `users`

### Key Field Evolution

`executions`:

- Retry: `retry_limit` / `retry_count` / `retry_delay_seconds` / `retry_of_execution_id` / `retried_at`
- Resource limits: `cpu_cores` / `memory_mb` / `timeout_seconds`
- Version and registry credentials: `spider_version` / `registry_auth_ref`

`scheduled_tasks`:

- Schedule cursor: `last_materialized_at`
- Retry strategy: `retry_limit` / `retry_delay_seconds`
- Version and registry credentials: `spider_version` / `registry_auth_ref`

`spider_versions`:

- `(spider_id, version)` unique constraint
- `registry_auth_ref`

### Key Indexes (examples)

- `idx_executions_status_created_at`
- `idx_executions_spider_version`
- `idx_executions_registry_auth_ref`
- `idx_scheduled_tasks_enabled`
- `idx_scheduled_tasks_spider_version`
- `idx_scheduled_tasks_registry_auth_ref`
- `idx_spider_versions_spider_version`
- `idx_spider_versions_registry_auth_ref`
- `idx_node_heartbeats_node_id_seen_at`
- `idx_alert_rules_enabled_type`
- `idx_alert_events_created_at`

## MongoDB

Collection: `execution_logs`

Typical document:

```json
{
  "_id": "uuid",
  "execution_id": "exec-id",
  "message": "log text",
  "created_at": "2026-05-02T13:00:00Z"
}
```

Recommended index: `(execution_id, created_at)`.

## Redis

### Node Online State

- `nodes:<node-id>`: Node state JSON (with TTL)
- `nodes:online`: Set of online nodes

### Execution Queue

- `executions:pending`
- `executions:inflight`

The claim flow depends on Redis atomic operations to prevent duplicate claiming by multiple agents.
