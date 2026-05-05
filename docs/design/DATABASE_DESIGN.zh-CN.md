# 数据库设计文档（2026-05 快照）

## 概述

平台使用三种存储：

- PostgreSQL：业务主数据与告警事件
- MongoDB：执行日志
- Redis：执行队列与在线节点临时状态

## PostgreSQL

迁移目录：`deploy/migrations/postgres`

当前迁移（按文件名顺序）：

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

### 核心表

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

### 关键字段演进

`executions`：

- 重试：`retry_limit` / `retry_count` / `retry_delay_seconds` / `retry_of_execution_id` / `retried_at`
- 资源限制：`cpu_cores` / `memory_mb` / `timeout_seconds`
- 版本与仓库凭据：`spider_version` / `registry_auth_ref`

`scheduled_tasks`：

- 调度游标：`last_materialized_at`
- 重试策略：`retry_limit` / `retry_delay_seconds`
- 版本与仓库凭据：`spider_version` / `registry_auth_ref`

`spider_versions`：

- `(spider_id, version)` 唯一约束
- `registry_auth_ref`

### 关键索引（示例）

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

集合：`execution_logs`

典型文档：

```json
{
  "_id": "uuid",
  "execution_id": "exec-id",
  "message": "log text",
  "created_at": "2026-05-02T13:00:00Z"
}
```

建议索引：`(execution_id, created_at)`。

## Redis

### 节点在线态

- `nodes:<node-id>`：节点状态 JSON（带 TTL）
- `nodes:online`：在线节点集合

### 执行队列

- `executions:pending`
- `executions:inflight`

claim 流程依赖 Redis 原子操作，避免多 agent 重复领取。
