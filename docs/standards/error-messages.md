# 错误消息规范

> 适用范围：gateway、后端 service、前端错误展示、API 文档。

## 响应结构

所有 HTTP JSON 错误响应保持统一结构：

```json
{ "error": "message" }
```

暂不引入多字段错误结构，避免前后端在当前阶段产生额外契约复杂度。

## 状态码约定

| 状态码 | 用途 |
| --- | --- |
| `400` | 请求格式、参数、校验失败 |
| `401` | 未认证、凭证缺失或无效 |
| `403` | 已认证但无权限 |
| `404` | 资源不存在、API 版本或路由不支持 |
| `409` | 唯一性冲突、状态冲突 |
| `429` | gateway 限流 |
| `500` | 服务内部未知错误 |
| `502` | gateway 无法访问上游服务 |

## 消息稳定性

1. `error` 字符串视为前端可映射语义，不应随意改动。
2. 新错误消息应简短、英文、小写为主。
3. 不在错误消息里泄露 token、密码、连接串、数据库内部错误。
4. 面向用户的本地化文案由前端 i18n 负责。
5. 后端日志可以记录更多上下文，但响应体只返回稳定语义。

## Gateway 错误语义

| 场景 | 状态码 | error |
| --- | --- | --- |
| 缺少 Bearer Token | `401` | `missing bearer token` |
| Authorization 方案错误 | `401` | `invalid authorization scheme` |
| Bearer Token 无效 | `401` | `invalid bearer token` |
| JWT secret 未配置 | `401` | `jwt secret is not configured` |
| internal token 未配置 | `401` | `internal token is not configured` |
| internal token 错误 | `401` | `unauthorized internal route` |
| 请求被限流 | `429` | `rate limit exceeded` |
| 未知上游服务 | `502` | `unknown upstream service` |
| 上游地址非法 | `500` | `invalid upstream service url` |
| 上游不可用 | `502` | `upstream service unavailable` |
| API 版本或路由不支持 | `404` | `unsupported api version or route` |
| 非 API 路由不存在 | `404` | `route not found` |

## Service 错误建议

| 场景 | 状态码 | error 示例 |
| --- | --- | --- |
| 资源不存在 | `404` | `project not found` |
| 创建重名资源 | `409` | `spider version already exists` |
| 参数缺失 | `400` | `projectId is required` |
| 状态不允许 | `409` | `execution cannot be started from current status` |
| 外部连接失败 | `400` 或 `502` | `datasource connection failed` |
| 告警 Webhook 失败 | `502` | `webhook delivery failed` |

## 前端展示原则

1. 优先按 `error` 原文映射到 i18n key。
2. 未映射错误展示通用文案加原始消息，例如“操作失败：xxx”。
3. 权限、认证错误应引导用户重新登录或检查权限。
4. 限流错误应提示稍后重试。
5. 上游不可用错误应提示服务暂不可用，而不是暴露内部服务名。
