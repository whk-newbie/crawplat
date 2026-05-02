# 开发路线图

## 概述

Crawler Platform 是一个分布式爬虫管理平台，采用微服务架构，分阶段迭代开发。本文档记录各阶段的目标、范围和当前进度。

---

## Phase 1 - MVP 基础 (已完成)

**目标：** 搭建项目骨架，验证微服务架构和 Compose 编排可行性。

**交付内容：**

- [x] monorepo 目录结构 + go.work 多模块管理
- [x] Gateway 反向代理骨架
- [x] iam-service 登录和 JWT 签发
- [x] project-service 项目创建/列表
- [x] spider-service 爬虫创建/列表
- [x] node-service 节点心跳
- [x] execution-service 执行记录（内存）
- [x] datasource-service 数据源配置
- [x] Docker Compose 编排（postgres + redis + mongo + 全部服务）
- [x] Web 前端 Vue 3 Shell
- [x] smoke test 脚本

---

## Phase 2 - 执行持久化和真实运行 (已完成)

**目标：** 将核心数据从内存存储迁移到真实数据库，实现端到端的爬虫执行流水线。

**交付内容：**

- [x] PostgreSQL 持久化核心元数据（projects / spiders / datasources / executions）
- [x] MongoDB 持久化执行日志
- [x] Redis 队列协调执行任务分发（pending -> inflight）
- [x] Redis 存储节点心跳和在线集合
- [x] Agent 实现 Docker 容器运行时（docker_runner.go）
- [x] Agent 轮询执行队列，完成 claim -> start -> log -> complete/fail 全流程
- [x] 内部执行 API（/internal/v1/executions/*）+ X-Internal-Token 认证
- [x] Web 前端实现 Executions / ExecutionDetail / Spiders 页面
- [x] Go/Python 示例爬虫镜像（go-echo / python-echo）
- [x] 数据库迁移框架（deploy/migrations/postgres/）
- [x] Phase 2 smoke checklist 验证通过

---

## Phase 3 - 调度、重试和监控 (已完成)

**目标：** 添加定时调度执行、失败重试和平台监控能力。

**交付内容：**

- [x] scheduler-service 调度服务
  - [x] Schedule CRUD（cron 表达式 + 重试策略）
  - [x] 定时轮询物化 scheduled execution
  - [x] 乐观锁防重复物化（last_materialized_at 游标）
  - [x] catch-up 支持（每轮最多物化 16 次）
- [x] 重试策略
  - [x] executions 表扩展重试字段（retry_limit / retry_count / retry_delay_seconds / retry_of_execution_id / retried_at）
  - [x] 重试物化端点（/internal/v1/executions/retries/materialize）
  - [x] scheduler-service 轮询驱动重试
- [x] monitor-service 监控服务
  - [x] 监控概览 API（/api/v1/monitor/overview）
  - [x] PostgreSQL 聚合执行状态计数
  - [x] Redis 查询在线节点计数
- [x] Web 前端
  - [x] SchedulesView 页面（创建 + 列表）
  - [x] MonitorView 页面（统计卡片 + 原始 JSON）
- [x] Docker Compose 开发工作流（docker-compose.dev.yml + air 热重载）
- [x] Phase 3 smoke checklist 验证通过

---

## Phase 4 - 节点资产和历史监控 (已完成)

**目标：** 持久化节点历史信息，支持离线节点统计和节点资产管理。

**交付内容：**

- [x] 节点信息持久化到 PostgreSQL（注册信息 + 历史在线记录）
- [x] 在线/离线节点统计（monitor overview 包含 offline 计数）
- [x] 节点详情 API（capabilities / 历史执行记录 / 在线时段 / heartbeat 历史）
- [x] 节点 sessions API（在线时段聚合 + 摘要统计）
- [x] 执行查询过滤（status / from / to / limit / offset）
- [x] Web 前端 Nodes 管理页面（列表 + 详情 + 心跳时间线 + sessions 表格）

---

## Phase 5 - 前端重构 (已完成)

**目标：** 引入 Element Plus，将前端从 MVP Shell 重构为完整的后台管理系统。

**交付内容：**

- [x] 引入 Element Plus 全量导入配置
- [x] 重构整体布局（ElContainer + ElMenu 侧边栏 + ElHeader）
- [x] 完善 LoginView（ElForm + ElInput + ElButton + ElMessage 反馈）
- [x] 完善 ProjectsView（ElTable + ElDialog 创建表单 + ElEmpty）
- [x] 完善 SpidersView（ElTable + ElDialog + ElSelect + ElTag）
- [x] 完善 ExecutionsView（ElCard + ElInput + ElDialog）
- [x] 完善 ExecutionDetailView（ElDescriptions + ElTimeline + ElPageHeader + ElTag 状态映射）
- [x] 完善 SchedulesView（ElTable + ElDialog + ElInputNumber + ElSwitch）
- [x] 完善 MonitorView（ElStatistic 卡片 + ElCollapse 原始 JSON + ElAlert）
- [x] 完善 NodesView（ElTable + ElDescriptions + ElTimeline + ElDatePicker + ElStatistic sessions）
- [x] 完善 DatasourcesView（ElTable + ElDialog 动态配置行 + 测试/预览操作结果展示）
- [x] 全局错误处理使用 ElMessage
- [x] 全部 9 个测试文件通过

---

## Phase 6 - API 网关增强 (已完成)

**目标：** Gateway 从简单反向代理升级为真正的 API 网关。

**交付内容：**

- [x] JWT 认证中间件（Gateway 层统一校验）
- [x] 请求速率限制（可配置 rate limiter）
- [x] 请求日志和链路追踪（request ID + access logging）
- [x] API 版本管理（/api/v1 路由）
- [x] 服务发现（service registry + 可配置 upstream 覆盖）
- [x] Web 静态资源 auth guard

---

## Phase 7 - 高级功能 (进行中)

**目标：** 平台功能完善，面向生产场景。

**规划内容：**

- [ ] 多租户隔离
- [ ] 高级 RBAC（角色 + 权限 + 团队）
- [ ] 任务 DAG / 工作流编排
- [x] 爬虫版本管理（版本持久化 + 版本 API + 前端版本管理）
- [ ] 镜像仓库集成
- [x] 执行资源限制（CPU / Memory / 超时）
- [x] 告警规则和通知（执行失败、节点离线，含 webhook 事件持久化）
- [x] 数据源真实连接测试和预览（替换当前 mock）
- [ ] 生产部署方案（Kubernetes / Helm）

---

## 已知技术债务

| 项 | 所在阶段 | 说明 |
|----|---------|------|
| IAM 用户内存存储 | Phase 1 | 用户存储在内存 map 中，重启丢失 |
| 密码明文比较 | Phase 1 | 未做 hash 处理 |
| ~~数据源 test/preview mock~~ | ~~Phase 1~~ | ✅ Phase 7 已替换为真实连接探测与预览 |
| 列表 API 能力不完整 | Phase 2 | 多数列表已支持分页；execution 列表查询仍待补齐 |
| ~~Gateway JWT 未强制~~ | ~~Phase 1~~ | ✅ Phase 6 已实现 Gateway JWT 中间件 |
| ~~前端部分页面仍为 placeholder~~ | ~~Phase 1~~ | ✅ Phase 5 已完成全部 9 个页面重构 |
| ~~节点 offline 统计缺失~~ | ~~Phase 3~~ | ✅ Phase 4 已实现节点持久化和离线统计 |
| ~~服务地址硬编码~~ | ~~Phase 1~~ | ✅ Phase 6 已实现 service registry |
