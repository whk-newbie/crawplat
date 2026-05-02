# 前端设计文档（2026-05 快照）

## 技术栈

- Vue 3 + TypeScript
- Vue Router
- Pinia
- Element Plus
- Vite
- Vitest

## 页面与路由

- `/login` 登录
- `/projects` 项目管理
- `/spiders` 爬虫与版本管理
- `/executions` 手动执行创建
- `/executions/:id` 执行详情与日志
- `/schedules` 调度任务管理
- `/datasources` 数据源管理
- `/monitor` 监控与告警
- `/nodes` 节点资产与历史

## 当前实现状态

所有上述页面均已落地，不再存在 placeholder 页面。

## 关键交互

### Spiders

- 支持创建 spider
- 支持版本列表与版本创建
- 版本可设置 `registryAuthRef`
- 支持“Load Refs”拉取项目级 registry auth refs

### Executions

- 支持按项目/爬虫/版本创建执行
- 支持 `registryAuthRef`、资源限制字段
- `registryAuthRef` 输入支持可选下拉（可创建新值）

### Schedules

- 支持 cron、启停、重试策略
- 支持 `spiderVersion` + `registryAuthRef`
- `registryAuthRef` 同样支持加载项目 refs

### Monitor

- overview 统计卡片
- 告警规则创建/列表/更新
- 告警事件分页列表

### Nodes

- 节点列表分页
- 节点详情（最近 heartbeat、执行历史筛选）
- sessions 聚合视图

### Datasources

- 数据源创建与列表
- test/preview 真实探测结果展示

## API 层

目录：`apps/web/src/api/*`

- 所有请求通过 `client.ts` 统一封装
- 默认走 gateway 的 `/api/v1/*`
- 列表接口统一返回 `{ items, total, limit, offset }` 的分页结构（若后端支持）

## 测试

- 视图级测试位于 `apps/web/src/views/__tests__`
- 当前包含 projects/spiders/executions/schedules/datasources/monitor/nodes/login 等关键页面测试
