# 平台术语表

> 适用范围：前端文案、API 文档、设计文档、注释和验收清单。
>
> 原则：中文面向产品语义，英文面向代码/API 语义。代码标识符保持英文，用户可见文案按当前语言展示。

## 核心术语

| 英文术语 | 中文译法 | 说明 | 代码/API 建议 |
| --- | --- | --- | --- |
| Project | 项目 | 业务隔离与资源组织单元 | `project`, `projectId` |
| Spider | 爬虫 | 可执行的采集任务定义 | `spider`, `spiderId` |
| Spider Version | 爬虫版本 | 爬虫镜像、参数与版本元数据的快照 | `spiderVersion`, `spiderVersionId` |
| Execution | 执行 | 一次爬虫运行实例 | `execution`, `executionId` |
| Execution Log | 执行日志 | 执行期间产生的日志记录 | `executionLog`, `logs` |
| Schedule | 调度 | 定时或周期性触发执行的配置 | `schedule`, `scheduleId` |
| Node | 节点 | 运行 agent 的计算节点 | `node`, `nodeId` |
| Node Session | 节点会话 | 节点连续在线时段 | `session`, `sessions` |
| Datasource | 数据源 | 被爬虫或平台访问的外部数据连接配置 | `datasource`, `datasourceId` |
| Monitor | 监控 | 平台状态聚合与告警入口 | `monitor`, `overview` |
| Alert Rule | 告警规则 | 定义何时触发告警的规则 | `alertRule`, `ruleId` |
| Alert Event | 告警事件 | 已触发并持久化的告警记录 | `alertEvent`, `eventId` |
| Registry Auth Ref | 镜像仓库凭据引用 | 指向镜像仓库认证配置的逻辑引用 | `registryAuthRef` |
| Retry | 重试 | 对失败执行进行重新物化或重新运行 | `retry`, `retryLimit`, `retryCount` |
| Resource Limit | 资源限制 | CPU、内存、超时等执行约束 | `cpuCores`, `memoryMB`, `timeoutSeconds` |
| Request ID | 请求 ID | gateway 生成或透传的请求追踪标识 | `X-Request-ID`, `requestID` |
| Internal Token | 内部令牌 | 内部 API 的服务间调用凭证 | `X-Internal-Token` |

## 状态术语

| 英文值 | 中文展示 | 适用对象 |
| --- | --- | --- |
| pending | 等待中 | execution |
| running | 运行中 | execution |
| succeeded | 成功 | execution |
| failed | 失败 | execution / alert event |
| online | 在线 | node |
| offline | 离线 | node |
| enabled | 已启用 | schedule / alert rule |
| disabled | 已禁用 | schedule / alert rule |
| firing | 触发中 | alert event |
| resolved | 已恢复 | alert event |

## 使用规则

1. 用户可见文案优先使用中文译法或英文资源文件中的对应英文。
2. API 字段、环境变量、配置 key 不翻译。
3. 文档首次出现重要英文术语时，可写作“中文译法（English Term）”。
4. 同一个页面、接口、文档内不得混用多个中文译法。
5. 新增领域术语必须先补充到本表，再进入前端资源、API 文档或设计文档。
