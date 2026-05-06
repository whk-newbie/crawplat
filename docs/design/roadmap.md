# Crawler Platform 路线图

## Phase 1-6: MVP 基础（已完成）

基础设施、IAM 登录、project/spider/execution/node/datasource 服务、Gateway 网关、Web Shell、Docker Compose 编排。

详见 `docs/architecture/mvp-progress-2026-04-12.zh-CN.md`。

## Phase 7: 多租户隔离（已完成）

**目标**：所有服务 API 支持 organization 级别的数据隔离，通过 `X-Org-ID` 请求头传递租户信息。

**实现模式**（Router → Service → Repo 三层）：

```
Router:  orgID := c.GetHeader("X-Org-ID")
Service: func Create(orgID string, ...)
Repo:    WHERE ($1 = '' OR organization_id = $1)
```

- 空 `orgID` 保持向后兼容（返回所有记录）
- 内部 API（`X-Internal-Token` 认证）传空 orgID，不做隔离

**已完成的实现**：

| 服务 | Router | Service | Repo | 测试 |
|------|--------|---------|------|------|
| datasource-service | ✓ | ✓ | ✓ | ✓ |
| project-service | ✓ | ✓ | ✓ | ✓ |
| spider-service | ✓ | ✓ | ✓ | ✓ |
| execution-service | ✓ | ✓ | ✓ | ✓ |
| scheduler-service | ✓ | ✓ | ✓ | ✓ |
| node-service | ✓ | ✓ | ✓ | ✓ |
| monitor-service | ✓ | ✓ | ✓ | ✓ |

**数据模型变更**：所有业务表均添加 `organization_id` 列。

**提交**: `f73a8be` feat: add organization-scoped filtering across all services

## Phase 8: 执行引擎完善（待开发）

- [ ] 调度触发执行（cron → CreateExecution）
- [ ] 执行重试逻辑完善
- [ ] 执行超时处理
- [ ] Agent 端容器生命周期管理

## Phase 9: 监控告警完善（待开发）

- [ ] 告警规则 CRUD UI
- [ ] 告警事件列表与搜索
- [ ] Webhook 投递重试与状态追踪
- [ ] 告警统计仪表板

## Phase 10: UI 完善（待开发）

- [ ] Spider 管理页面（CRUD）
- [ ] 数据源管理页面（连接测试、预览）
- [ ] 调度管理页面
- [ ] 节点管理页面
- [ ] 组织管理页面（多租户后台）

## Phase 11: 生产就绪（待开发）

- [ ] 数据库迁移工具集成
- [ ] 日志聚合（结构化日志）
- [ ] 指标采集（Prometheus）
- [ ] CI/CD 流水线
- [ ] 安全审计（API 访问日志）
