# 服务间通信设计（2026-05 快照）

## 总体拓扑

- 浏览器 -> web（Vite/Nginx）
- 浏览器 -> gateway（`/api/v1/*`）
- gateway -> 各业务服务（reverse proxy）
- agent 直连 node-service 与 execution-service 内部接口
- scheduler-service 调用 execution-service 公共/内部接口
- monitor-service 聚合 PostgreSQL + Redis

## 通信与认证矩阵

| 调用方 | 被调用方 | 路径 | 认证 |
|---|---|---|---|
| Browser | Gateway | `/api/v1/*` | Bearer JWT（可配置强制） |
| Gateway | iam-service | `/api/v1/auth/*` | 透传 |
| Gateway | project-service | `/api/v1/projects*` | 透传 |
| Gateway | spider-service | `/api/v1/projects/:id/spiders*`, `/api/v1/spiders/:id/versions*` | 透传 |
| Gateway | execution-service | `/api/v1/executions*` | 透传 |
| Gateway | execution-service | `/internal/v1/executions*` | `X-Internal-Token` |
| Gateway | node-service | `/api/v1/nodes*` | 透传 |
| Gateway | datasource-service | `/api/v1/datasources*` | 透传 |
| Gateway | scheduler-service | `/api/v1/schedules*` | 透传 |
| Gateway | monitor-service | `/api/v1/monitor*` | 透传 |
| Agent | node-service | `/api/v1/nodes/:id/heartbeat` | 无 |
| Agent | execution-service | `/internal/v1/executions/*` | `X-Internal-Token` |
| Scheduler | execution-service | `/api/v1/executions` | 无（内网调用） |
| Scheduler | execution-service | `/internal/v1/executions/retries/materialize` | `X-Internal-Token` |

## 执行链路

1. 用户调用 `POST /api/v1/executions` 创建执行（写 PG + 入 Redis pending）。
2. agent 轮询 `POST /internal/v1/executions/claim` 领取任务。
3. agent 调 `start`，执行容器并追加日志到 `/logs`。
4. 执行结束后调用 `complete` 或 `fail`。
5. scheduler 周期调用 `retries/materialize` 物化重试任务。

## 调度链路

1. scheduler 扫描启用任务并计算 cron。
2. 使用 `last_materialized_at` 游标推进并防重。
3. 到触发窗口后调用 execution create API 生成执行。

## 告警链路

1. monitor-service 轮询失败执行与离线节点。
2. 按规则类型匹配 `alert_rules`。
3. 发送 webhook，结果落表 `alert_events`。
4. 前端通过 `/api/v1/monitor/alerts/events` 分页查询历史事件。

## 私有镜像拉取链路

1. 执行记录可带 `registryAuthRef`（或从 spider version 继承）。
2. agent 从 `IMAGE_REGISTRY_AUTH_MAP` 读取 registry 凭据映射（支持 host-key 与 named-ref+server）。
3. 当镜像 host 匹配映射项，或 `registryAuthRef` 命中映射项时，先 `docker login`/`docker pull` 再 `docker run`。
