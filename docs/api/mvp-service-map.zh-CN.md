# MVP 服务映射表

## 服务职责

- `gateway` -> 公共入口与请求代理
- `iam-service` -> 认证
- `project-service` -> 项目创建 / 列表（PostgreSQL）
- `spider-service` -> 爬虫创建 / 列表（PostgreSQL）
- `execution-service` -> 执行元数据（PostgreSQL）、日志（MongoDB）、队列驱动的生命周期转换
- `node-service` -> 节点心跳（Redis）
- `datasource-service` -> 数据源配置与预览
- `agent` -> 心跳、执行轮询、Docker 运行时执行

## 关键路由

### `gateway`

- `ANY /api/v1/auth`
- `ANY /api/v1/auth/*path`
- `ANY /api/v1/projects`
- `ANY /api/v1/projects/:projectId/spiders`
- `ANY /api/v1/executions`
- `ANY /api/v1/executions/*path`
- `ANY /api/v1/nodes`
- `ANY /api/v1/nodes/*path`
- `ANY /api/v1/datasources`
- `ANY /api/v1/datasources/*path`
- `ANY /internal/v1/executions/claim`
- `ANY /internal/v1/executions/:id/start`
- `ANY /internal/v1/executions/:id/logs`
- `ANY /internal/v1/executions/:id/complete`
- `ANY /internal/v1/executions/:id/fail`

### `iam-service`

- `POST /api/v1/auth/login`

### `project-service`

- `POST /api/v1/projects`
- `GET /api/v1/projects`

### `spider-service`

- `POST /api/v1/projects/:projectId/spiders`
- `GET /api/v1/projects/:projectId/spiders`

### `execution-service`

- `POST /api/v1/executions`
- `POST /api/v1/executions/:id/logs`
- `GET /api/v1/executions/:id`
- `GET /api/v1/executions/:id/logs`
- `POST /internal/v1/executions/claim`
- `POST /internal/v1/executions/:id/start`
- `POST /internal/v1/executions/:id/logs`
- `POST /internal/v1/executions/:id/complete`
- `POST /internal/v1/executions/:id/fail`

### `node-service`

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes`

### `datasource-service`

- `POST /api/v1/datasources`
- `GET /api/v1/datasources`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`

## 说明

- 网关是 Web 应用和外部调用方唯一可访问的公共 API 接口面。
- 内部执行路由仅供执行 Worker 使用，需要 `X-Internal-Token`。
- Agent 直接使用节点心跳路由保持存活状态，并通过 execution-service 内部路由完成领取 / 启动 / 日志 / 完成 / 失败流程。
