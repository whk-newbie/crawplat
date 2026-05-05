# Phase 3 冒烟检查清单

## 目标

在可正常运行的 Phase 2 执行流水线基础上，验证 Phase 3 新增功能：

- `scheduler-service` 已加入 Compose 服务栈
- 可通过网关创建和查询调度
- 调度执行可将重试配置传递到执行记录
- 重试物化端点可从失败执行创建重试执行
- 监控概览路由和 Web 页面可正常访问

## 前置条件

- Docker 可用
- 示例镜像已存在：
  - `crawler/go-echo:latest`
  - `crawler/python-echo:latest`

## 启动

```bash
make migrate
make up
```

## 调度 API

先创建项目和爬虫，然后通过网关创建调度：

```bash
curl -sS -X POST http://localhost:8080/api/v1/schedules \
  -H 'Content-Type: application/json' \
  -d '{
    "projectId":"<project-id>",
    "spiderId":"<spider-id>",
    "name":"nightly-go-echo",
    "cronExpr":"*/5 * * * *",
    "enabled":true,
    "image":"crawler/go-echo:latest",
    "command":["./go-echo"],
    "retryLimit":2,
    "retryDelaySeconds":30
  }'
```

预期结果：

- `201 Created`
- 响应中包含 `retryLimit: 2`
- 响应中包含 `retryDelaySeconds: 30`

查询调度列表：

```bash
curl -sS http://localhost:8080/api/v1/schedules
```

预期结果：

- 列表中显示刚创建的调度

## 重试物化

此端点为内部接口，需要 `X-Internal-Token`。

```bash
curl -i -X POST http://localhost:8085/internal/v1/executions/retries/materialize \
  -H 'X-Internal-Token: change-me'
```

预期结果：

- 存在可重试的失败执行时返回 `201 Created`
- 无可重试的失败执行时返回 `204 No Content`

如果返回 `201 Created`，确认返回的执行记录包含：

- `triggerSource: "retry"`
- `retryCount` 已递增
- `retryOfExecutionId` 指向失败的原始执行

## 监控路由

```bash
curl -sS http://localhost:8080/api/v1/monitor/overview
```

预期结果：

- `200 OK`
- JSON 结构中包含 `executions` 和 `nodes`
- 执行计数器反映 `executions` 表中当前存储的记录
- 节点计数器反映当前 Redis 心跳集合

注意：

- MVP 中离线节点数保持 `0`，因为仅持久化在线心跳

## Web 路由

打开：

- `http://localhost:3000/monitor`

预期结果：

- 页面正常加载
- 刷新按钮可用
- 原始数据区域正确渲染监控概览响应

## 清理

```bash
make down
```
