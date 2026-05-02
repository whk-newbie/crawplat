# API 设计文档（2026-05 快照）

## 概述

Crawler Platform API 分为两类：

- 公共 API：`/api/v1/*`，由 gateway 统一暴露给 Web 和外部客户端。
- 内部 API：`/internal/v1/executions/*`，用于 agent / scheduler 与 execution-service 交互。

所有接口使用 JSON。

## 认证与网关策略

- JWT：gateway 已实现 Bearer Token 校验（`GATEWAY_ENFORCE_JWT` 可配置）。
- 内部令牌：`X-Internal-Token`，值来自 `INTERNAL_API_TOKEN`（未设置时回退 `JWT_SECRET`）。
- 网关增强：
  - rate limit（`GATEWAY_RATE_LIMIT_*`）
  - request-id 与 access logging
  - API 版本路由（默认 `v1`，可通过 `GATEWAY_API_SUPPORTED_VERSIONS` 扩展）

## 通用响应约定

错误统一为：

```json
{ "error": "message" }
```

常见状态码：`200`、`201`、`204`、`400`、`401`、`404`、`409`、`500`、`502`。

## 公共 API 清单

### iam-service

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### project-service

- `POST /api/v1/projects`
- `GET /api/v1/projects?limit=<n>&offset=<n>`

### spider-service

- `POST /api/v1/projects/:projectId/spiders`
- `GET /api/v1/projects/:projectId/spiders?limit=<n>&offset=<n>`
- `GET /api/v1/projects/:projectId/registry-auth-refs`
- `POST /api/v1/spiders/:spiderId/versions`
- `GET /api/v1/spiders/:spiderId/versions`

### execution-service

- `POST /api/v1/executions`
- `GET /api/v1/executions/:id`
- `GET /api/v1/executions/:id/logs`
- `POST /api/v1/executions/:id/logs`

`POST /api/v1/executions` 支持：

- `spiderVersion`
- `registryAuthRef`
- 资源限制：`cpuCores` / `memoryMB` / `timeoutSeconds`
- 重试字段：`retryLimit` / `retryCount` / `retryDelaySeconds` / `retryOfExecutionId`

当请求未显式提供 `registryAuthRef` 时，execution-service 会尝试从指定 spider version 继承。

### scheduler-service

- `POST /api/v1/schedules`
- `GET /api/v1/schedules?limit=<n>&offset=<n>`

调度任务支持：`spiderVersion`、`registryAuthRef`、重试策略字段。

### node-service

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes?limit=<n>&offset=<n>`
- `GET /api/v1/nodes/:id`
- `GET /api/v1/nodes/:id/sessions?limit=<n>&gapSeconds=<n>`

`GET /api/v1/nodes/:id` 支持执行历史过滤参数：

- `executionLimit` / `executionOffset`
- `executionStatus`
- `executionFrom` / `executionTo`（RFC3339）

### datasource-service

- `POST /api/v1/datasources`
- `GET /api/v1/datasources?projectId=<id>&limit=<n>&offset=<n>`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`

`test/preview` 现为真实探测，不再是固定 mock 响应。

### monitor-service

- `GET /api/v1/monitor/overview`
- `POST /api/v1/monitor/alerts/rules`
- `GET /api/v1/monitor/alerts/rules`
- `PATCH /api/v1/monitor/alerts/rules/:id`
- `GET /api/v1/monitor/alerts/events?limit=<n>&offset=<n>`

告警规则支持两类：执行失败、节点离线。告警事件会持久化到 PostgreSQL。

## 内部 API（execution-service）

以下接口由 gateway 和 execution-service 双重校验内部令牌：

- `POST /internal/v1/executions/claim`
- `POST /internal/v1/executions/:id/start`
- `POST /internal/v1/executions/:id/logs`
- `POST /internal/v1/executions/:id/complete`
- `POST /internal/v1/executions/:id/fail`
- `POST /internal/v1/executions/retries/materialize`

## 与私有镜像仓库的接口协同

- `registryAuthRef` 在 spider versions / schedules / executions 三处可传递。
- agent 启动时读取 `IMAGE_REGISTRY_AUTH_MAP`：
  - 支持 host-key：`"<registry-host>" -> credential`
  - 支持 named-ref：`"<registryAuthRef>" -> { server: "<registry-host>", ... }`
- 当匹配镜像 registry host，或 execution 的 `registryAuthRef` 命中映射时，执行 `docker login` + `docker pull`。
