# MVP 进度更新

日期：2026-04-12
分支：`feat/mvp-foundation`
Worktree：`/home/iambaby/goland_projects/crawler-platform/.worktrees/mvp-foundation`

## 摘要

MVP 基础实现在代码和文档层面已完成。全部 12 项计划任务现已实现并通过审查：

- 仓库骨架
- 共享 Go 基础包
- IAM 登录模块
- project-service CRUD 模块
- spider-service 创建 / 列表模块
- node-service 与 Agent 心跳循环
- execution-service 手动执行与日志采集模块
- datasource-service CRUD、连接测试与预览模块
- 网关路由与鉴权透传
- Vue Web MVP Shell
- Docker Compose 服务栈编排与冒烟流程
- 入门与架构文档

## 已完成的组件

### 仓库基础

- 根级 `Makefile`
- `go.work`
- 根级 `go.mod`
- `.env.example`
- 初始 Docker Compose 骨架
- 通过根级 `make test` 执行动态嵌套模块测试

### 共享 Go 包

- JWT 签发 / 解析工具
- config 结构体
- 轻量 HTTP 响应工具
- 轻量 PostgreSQL / Redis / Mongo 封装

### IAM 服务

- `POST /api/v1/auth/login`
- 通过显式环境变量开关控制的预置管理员路径
- 启动时要求配置 JWT secret
- 覆盖成功和异常路径的 HTTP 层和 Service 层测试

### Project 服务

- `POST /api/v1/projects`
- `GET /api/v1/projects`
- 内存项目存储
- JSON 响应结构测试
- Service 和 Router 层覆盖

### Spider 服务

- `POST /api/v1/projects/:projectId/spiders`
- `GET /api/v1/projects/:projectId/spiders`
- language / runtime 校验
- 项目维度列表行为
- Service 和 Router 层覆盖

### Node 服务与 Agent

- `POST /api/v1/nodes/:id/heartbeat`
- `GET /api/v1/nodes`
- 内存节点心跳追踪
- Agent 启动心跳
- 心跳失败异常暴露
- 优雅的信号驱动关闭
- Router 和 Agent 心跳测试

### Execution 服务

- `POST /api/v1/executions`
- `POST /api/v1/executions/:id/logs`
- `GET /api/v1/executions/:id`
- `GET /api/v1/executions/:id/logs`
- 以 `pending` 状态创建手动执行记录
- 内存执行和日志存储
- Service 和 Router 层乐观路径覆盖

### Datasource 服务

- `POST /api/v1/datasources`
- `GET /api/v1/datasources`
- `POST /api/v1/datasources/:id/test`
- `POST /api/v1/datasources/:id/preview`
- 对 `mongodb`、`redis`、`postgresql` 的数据源类型校验
- 带边界限制的只读预览行为
- Service 和 Router 层覆盖

### Gateway

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
- 受保护路由的 JWT 透传
- 路由级代理覆盖

### Web MVP Shell

- Vue 3 + TypeScript + Vite Shell
- 登录页与 Token 持久化
- 仪表盘摘要
- 项目视图
- 爬虫视图
- 执行视图
- 数据源视图
- Vitest 覆盖并纳入根级 `make test`

### Compose 与冒烟流程

- Go 服务和 Web Shell 的 Dockerfile
- `deploy/docker-compose/docker-compose.mcp.yml`（含数据存储和服务容器）
- `make up` 构建并启动工作流
- `make down` 清理工作流
- `deploy/scripts/smoke-mvp.sh`
- `docs/product/mvp-smoke-checklist.md`

### 入门与架构文档

- 根级 `README.md`
- `docs/architecture/mvp-overview.md`
- `docs/api/mvp-service-map.md`

## 当前说明

- 代码在独立的 Git Worktree 中开发，尚未合并回项目主分支。
- 当前在 Task 12 完成后，仓库级别处于干净状态。
- `make test` 在已实现的 MVP 模块中全部通过。

## Phase 2 补充

截至 2026-04-22，MVP 基础已超出原始 2026-04-12 快照：

- `project-service`、`spider-service` 和 `datasource-service` 的核心元数据现已持久化到 PostgreSQL。
- `node-service` 将心跳存活状态存储在 Redis 中。
- `execution-service` 将执行元数据存储在 PostgreSQL，执行日志存储在 MongoDB，队列状态存储在 Redis。
- 内部执行生命周期路由现覆盖领取、启动、追加日志、完成和失败等流程，受 `X-Internal-Token` 保护。
- 持久化和执行生命周期工作完成后 `make test` 依然全部通过。
- 由于当前环境不支持 Docker 桥接网络（`failed to add the host <=> sandbox pair interfaces: operation not supported`），完整的 `make up` 验证受阻。
