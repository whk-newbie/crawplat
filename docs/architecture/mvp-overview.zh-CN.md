# MVP 架构概览

## 当前状态

MVP 基础已实现为一个轻量级控制平面加执行平面 Agent：

- `gateway` 提供唯一的公共 API 入口。
- `iam-service` 处理登录和 JWT 签发。
- `project-service` 负责项目的创建和列表查询，数据持久化在 PostgreSQL。
- `spider-service` 负责项目维度的爬虫创建和列表查询，数据持久化在 PostgreSQL。
- `execution-service` 负责执行记录管理（PostgreSQL）、日志存储（MongoDB）和通过 Redis 协调执行队列。
- `node-service` 通过 Redis 追踪节点心跳和当前节点清单。
- `datasource-service` 负责数据源配置、连接检查和预览，数据持久化在 PostgreSQL。
- `agent` 运行在节点上，发送心跳、轮询执行任务，并将生命周期和日志更新回传给控制平面。

整个技术栈以 Docker Compose 优先设计。PostgreSQL、Redis、MongoDB 与 API 服务和 Web Shell 一起启动，使 MVP 环境与目标平台形态保持一致。当前阶段已在项目、爬虫、数据源、执行、日志、队列状态和节点存活等请求路径上使用了这些数据存储。

## 运行时形态

网关对外暴露 `/api/v1/*` 接口面，并将请求代理到对应的服务。各服务拥有自己的 Handler 并存储自己的领域状态，这使得实现可测试且即使在 MVP 阶段也能清晰展示服务边界。

网关同时暴露公共 `/api/v1/*` 接口面和部分内部 `/internal/v1/executions/*` 路由供执行 Worker 使用。内部执行路由受 `X-Internal-Token` 保护，而面向用户的请求继续走公共 API。

Agent 直接与节点服务通信以发送心跳更新。这使得节点存活状态独立于面向用户的 API 流量，并使执行平面拥有独立于控制平面的生命周期。执行 Worker 流程现已支持队列领取、启动、追加日志、完成和失败等状态转换，Agent 通过基于 Docker 的运行时接入该流程。

## 已实现内容

- 通过 `POST /api/v1/auth/login` 登录
- 项目创建 / 列表查询
- 项目维度爬虫创建 / 列表查询
- 执行生命周期：创建、获取、日志、领取、启动、完成和失败
- 基于 Redis 的节点心跳追踪和节点列表
- 数据源创建、列表、测试和预览
- 当前 API 接口面的网关路由
- 服务、数据存储、Web Shell、数据库迁移流程和支持 Docker 的 Agent 运行时的 Compose 编排
- 通过 Agent Poller 验证的 Go/Python Docker 执行路径

## MVP 尚未覆盖的内容

- 定时调度编排
- 任务 DAG 或工作流编排
- 多租户隔离
- 超出当前登录模块的高级 RBAC
- 超出基于 Compose 的 MVP 技术栈的生产加固
