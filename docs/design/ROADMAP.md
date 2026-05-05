# 开发路线图（Worktree 并行版）

> 说明
>
> - 本文档是 Crawler Platform 的**全新路线图**，用于指导后续按 **多个 worktree 并行开发** 的方式推进。
> - **不参考之前是否已经实现**，也**不沿用旧路线图中的完成状态**。
> - 仅依据当前设计文档与实现计划重写，目标是：**让团队可以同时开多个 worktree，并且尽量少发生互相阻塞**。
> - 路线图按**仓库 / 服务边界**切分，优先保证“一个 worktree 的改动范围尽量单一、可独立验收、可单独合并”。
> - 本路线图**不依赖 `roadmap.md` 之外的历史状态**，也不要求先清理旧计划。

---

## 1. 路线图设计原则

### 1.1 并行优先

路线图不是单线排期，而是“并行作战地图”。任务拆分以是否能放进独立 worktree 为首要标准，而不是以功能先后为首要标准。

可并行的工作单元通常满足以下条件：

- 修改目录边界尽量不重叠
- 对外契约明确或可提前冻结
- 依赖方向单向、稳定
- 验收方式独立
- 不需要频繁改同一批共享文件

### 1.2 以仓库为最小分派单元

本路线图默认将每个仓库 / 服务作为最小分派单元，并允许再向下拆分为“基础能力 worktree”“业务 worktree”“文档 worktree”“质量保障 worktree”。

原则上避免以下情况：

- 一个 worktree 同时改前端、后端、文档
- 一个 worktree 横跨多个服务的大量文件
- 一个 worktree 既做功能实现又做全量验收脚本

### 1.3 先冻结公共依赖，再并行展开

为了支持多个 worktree 同时开发，必须先冻结三类公共基线：

1. 术语与文案基线
2. 双语文档结构基线
3. 测试与验收基线

这些基线一旦稳定，后续各 worktree 就可以围绕自己的目录范围独立推进。

### 1.4 以“独立验收”作为完成标准

每个 worktree 的完成标准必须能单独验证，至少包括：

- 可以运行或编译
- 可以通过对应测试或检查
- 可以明确指出改了哪些目录
- 不依赖未合并的临时分支内容
- 与其他 worktree 的边界不冲突

---

## 2. Worktree 拓扑总览

为了支持并行开发，建议把整个项目拆成 4 类 worktree：

1. **公共规范 worktree**
   - 负责术语表、文案规则、双语文档模板、注释规范、验收模板。
2. **前端 worktree**
   - 负责国际化底座、通用组件、各业务页面双语化。
3. **后端 service worktree**
   - 负责各服务内的实现完善、接口一致性、中文注释补齐。
4. **文档与质量 worktree**
   - 负责 docs 双语化、CI 检查、smoke 回归。

这 4 类 worktree 的目标不是同时做完所有事情，而是让不同人能在不同 worktree 里**互不等待地推进**。

---

## 3. 推荐 worktree 分组

### 3.1 公共规范 worktree 组

#### `worktree-platform-spec`
负责统一平台术语、语言风格、中文注释规范、错误展示原则。

#### `worktree-test-harness`
负责测试模板、smoke checklist、回归检查标准。

#### `worktree-docs-bilingual-core`
负责双语文档结构、目录结构、文件命名和同步规则。

这三个 worktree 建议最先启动，因为它们会作为后续所有 worktree 的共同前提。

---

### 3.2 前端 worktree 组

#### `worktree-web-i18n-core`
负责前端国际化底座：语言切换、持久化、fallback、资源目录、通用 key 规范。

#### `worktree-web-shared-components`
负责布局、导航、面包屑、错误页、通知、弹窗、表单校验等通用组件国际化。

#### `worktree-web-auth-project`
负责登录、注册、项目管理页面的双语改造。

#### `worktree-web-spider-execution`
负责 Spider、Spider Version、Execution、Execution Logs 页面双语改造。

#### `worktree-web-schedule-node`
负责 Schedule、Node、会话和执行历史相关页面双语改造。

#### `worktree-web-datasource-monitor`
负责 Datasource、Monitor、告警规则和告警事件页面双语改造。

前端 worktree 的核心原则是：**先底座，再通用组件，再按业务域并行切页面**。

---

### 3.3 后端 service worktree 组

#### `worktree-gateway`
负责网关错误响应、鉴权边界、request-id、API 路由与中间件说明。

#### `worktree-iam-service`
负责登录 / 注册 / 认证错误映射和相关注释补齐。

#### `worktree-project-service`
负责项目 CRUD、分页、空列表、错误响应和核心注释。

#### `worktree-spider-service`
负责 Spider / Spider Version / registryAuthRef 传递与校验。

#### `worktree-execution-service`
负责执行创建、领取、启动、日志、完成、失败、重试和资源限制。

#### `worktree-scheduler-service`
负责调度物化、重试驱动、cron 触发、去重游标。

#### `worktree-node-service`
负责节点心跳、在线态、会话、执行过滤。

#### `worktree-datasource-service`
负责数据源 CRUD、test / preview、失败原因映射。

#### `worktree-monitor-service`
负责监控总览、告警规则、告警事件、Webhook 记录。

后端 worktree 的核心原则是：**每个服务只在自己的目录和契约范围内推进，不跨服务“顺手修”**。

---

### 3.4 文档与质量 worktree 组

#### `worktree-docs-product`
负责产品类文档双语化。

#### `worktree-docs-design`
负责设计类文档双语化。

#### `worktree-docs-architecture`
负责架构与通信类文档双语化。

#### `worktree-docs-howto`
负责开发、部署、运维类文档双语化。

#### `worktree-ci-lint`
负责注释规范检查、双语文档一致性检查、术语检查。

#### `worktree-e2e-smoke`
负责核心链路冒烟、双语切换冒烟、错误状态回归。

---

## 4. 里程碑划分

### Milestone 1 - 冻结公共基线

目标：把所有 worktree 都要共同遵守的东西先定下来。

包含 worktree：

- `worktree-platform-spec`
- `worktree-test-harness`
- `worktree-docs-bilingual-core`

完成后应输出：

- 统一术语表
- 中文注释规范
- 双语文档模板
- smoke / 回归模板

已冻结规范文档：

- `docs/standards/platform-terminology.md`
- `docs/standards/i18n-guidelines.md`
- `docs/standards/code-comments.md`
- `docs/standards/error-messages.md`
- `docs/standards/bilingual-docs.md`
- `docs/standards/worktree-acceptance.md`
- `docs/standards/smoke-checklist-template.md`

这一阶段完成前，不建议大规模启动业务 worktree。

---

### Milestone 2 - 建立前端可并行底座

目标：让前端具备语言系统和通用国际化能力。

包含 worktree：

- `worktree-web-i18n-core`
- `worktree-web-shared-components`
- `worktree-gateway`

完成后应输出：

- 前端语言切换能力
- 通用组件国际化
- 稳定的网关错误边界

这一阶段完成后，前端业务页面 worktree 可以并行展开。

---

### Milestone 3 - 后端服务并行落地

目标：各服务按仓库并行推进，实现接口稳定、核心逻辑清晰、注释补齐。

包含 worktree：

- `worktree-iam-service`
- `worktree-project-service`
- `worktree-spider-service`
- `worktree-execution-service`
- `worktree-scheduler-service`
- `worktree-node-service`
- `worktree-datasource-service`
- `worktree-monitor-service`

完成后应输出：

- 核心接口行为与设计文档一致
- 中文函数注释 / 文件注释补齐
- 关键错误响应稳定

这一阶段不要求所有前端页面已经完成，但要求后端契约尽量稳定，方便前端最后补齐。

---

### Milestone 4 - 前端业务页面并行收口

目标：把所有业务页面切到中英双语。

包含 worktree：

- `worktree-web-auth-project`
- `worktree-web-spider-execution`
- `worktree-web-schedule-node`
- `worktree-web-datasource-monitor`

完成后应输出：

- 核心页面双语可用
- 表单、弹窗、提示、错误页翻译完整
- 重要流程不依赖中文硬编码

---

### Milestone 5 - 文档与验收收口

目标：把文档、检查、冒烟测试收起来，形成可长期维护的体系。

包含 worktree：

- `worktree-docs-product`
- `worktree-docs-design`
- `worktree-docs-architecture`
- `worktree-docs-howto`
- `worktree-ci-lint`
- `worktree-e2e-smoke`

完成后应输出：

- 核心文档中英双语版本
- CI 约束与检查规则
- 冒烟回归脚本 / checklist

---

## 5. 详细 worktree 任务拆分

---

### 5.1 `worktree-platform-spec`

#### 目标

统一全平台术语、文案、注释、错误展示和双语维护规则，作为所有 worktree 的公共前提。

#### 任务

1. 建立术语表，冻结核心词汇译法：
   - Project
   - Spider
   - Spider Version
   - Execution
   - Schedule
   - Node
   - Datasource
   - Alert Rule
   - Alert Event
   - Registry Auth Ref
2. 定义前端 i18n key 命名规则。
3. 定义前端页面标题、菜单、按钮、状态的统一翻译原则。
4. 定义后端中文函数注释规范。
5. 定义后端中文文件注释规范。
6. 定义文档双语写作规范。
7. 定义错误消息展示与翻译映射原则。

#### 验收

- 术语表能直接被前端和文档引用。
- 规范内容足以支撑后续 worktree 不再反复讨论译法。

---

### 5.2 `worktree-test-harness`

#### 目标

建立统一的测试、回归和验收模板，避免每个 worktree 自己发明一套检查方式。

#### 任务

1. 建立 smoke checklist 模板。
2. 建立前端国际化测试模板。
3. 建立后端 API 回归模板。
4. 建立文档双语一致性检查模板。
5. 建立注释规范抽检模板。
6. 定义每类 worktree 的最小验收清单。

#### 验收

- 所有后续 worktree 都可以直接引用该模板。
- 检查项覆盖主流程、错误流程、国际化和文档一致性。

---

### 5.3 `worktree-docs-bilingual-core`

#### 目标

建立双语文档的目录结构、文件命名和同步规则。

#### 任务

1. 定义中文主文档与英文主文档的对应方式。
2. 定义共享术语表文档位置。
3. 定义标题层级同步规则。
4. 定义表格与接口示例同步规则。
5. 定义双语文件命名规范。
6. 定义未翻译内容的占位策略。

#### 验收

- 新文档知道该放哪里。
- 现有设计、产品、架构、How-to 文档可以套用同一套组织方式。

---

### 5.4 `worktree-web-i18n-core`

#### 目标

把前端国际化底座搭起来，为所有前端页面提供稳定语言能力。

#### 任务

1. 引入或确认 i18n 方案。
2. 建立 `zh-CN` 与 `en-US` 资源目录。
3. 建立语言切换状态管理。
4. 建立语言持久化策略。
5. 建立 fallback 规则。
6. 建立通用 key 命名规范。
7. 建立页面标题、菜单、按钮、提示统一读取方式。

#### 验收

- 切换语言后页面保持当前路由和状态。
- 缺失翻译 key 时不会崩溃。
- 底座可被业务页面 worktree 直接复用。

---

### 5.5 `worktree-web-shared-components`

#### 目标

把通用组件先国际化掉，减少后续业务页面重复劳动。

#### 任务

1. 国际化导航菜单。
2. 国际化面包屑。
3. 国际化错误页（401 / 403 / 404 / 500）。
4. 国际化通知、弹窗、确认框。
5. 国际化表单校验提示。
6. 国际化空状态、加载状态。
7. 国际化通用状态枚举展示。

#### 验收

- 通用组件在中英文环境下样式稳定。
- 业务页面无需重复实现同类翻译逻辑。

---

### 5.6 `worktree-web-auth-project`

#### 目标

完成登录、注册和项目管理的双语改造。

#### 任务

1. 登录页文案国际化。
2. 注册页文案国际化。
3. 登录失败 / 注册失败 / 权限不足提示国际化。
4. 项目列表页国际化。
5. 项目创建 / 编辑弹窗国际化。
6. 删除确认、空状态、表单校验国际化。
7. 菜单与页面标题国际化。

#### 验收

- 中英文切换下，登录、注册、项目管理主链路完整。
- 页面排版可承受英文长度。

#### 依赖

- `worktree-web-i18n-core`
- `worktree-web-shared-components`

---

### 5.7 `worktree-web-spider-execution`

#### 目标

完成 Spider、Spider Version、Execution、Execution Logs 的双语改造。

#### 任务

1. Spider 列表、创建、编辑页国际化。
2. Spider Version 列表、创建、选择器国际化。
3. Execution 列表、创建、详情页国际化。
4. Execution 日志页国际化。
5. 资源限制、重试参数、版本引用、仓库凭据引用文案国际化。
6. 状态枚举和错误信息映射国际化。

#### 验收

- 核心执行链路页面可中英切换。
- 长文案不破坏布局。
- 相关字段展示与设计文档一致。

#### 依赖

- `worktree-web-i18n-core`
- `worktree-web-shared-components`
- `worktree-spider-service`
- `worktree-execution-service`

---

### 5.8 `worktree-web-schedule-node`

#### 目标

完成 Schedule、Node、会话和执行历史页面的双语改造。

#### 任务

1. Schedule 列表、创建、编辑页国际化。
2. cron 输入提示和重试策略提示国际化。
3. Node 列表、详情、会话页国际化。
4. 在线 / 离线 / 健康状态提示国际化。
5. 执行历史过滤条件国际化。

#### 验收

- 调度和节点监控页面可中英切换。
- 状态、筛选、时间字段展示统一。

#### 依赖

- `worktree-web-i18n-core`
- `worktree-web-shared-components`
- `worktree-scheduler-service`
- `worktree-node-service`

---

### 5.9 `worktree-web-datasource-monitor`

#### 目标

完成 Datasource、Monitor、告警规则、告警事件页面的双语改造。

#### 任务

1. Datasource 列表、创建、编辑页国际化。
2. test / preview 操作结果国际化。
3. 数据源连接异常提示国际化。
4. Monitor 总览页国际化。
5. 告警规则列表 / 编辑 / 事件页国际化。
6. 告警状态、触发条件、Webhook 错误提示国际化。

#### 验收

- 数据源和监控页面能准确表达业务语义。
- 告警规则与事件状态名称统一。

#### 依赖

- `worktree-web-i18n-core`
- `worktree-web-shared-components`
- `worktree-datasource-service`
- `worktree-monitor-service`

---

### 5.10 `worktree-gateway`

#### 目标

把网关层稳定化、标准化，作为全系统统一入口和错误边界。

#### 任务

1. 统一错误响应结构。
2. 强化 JWT 校验与错误信息。
3. 强化内部令牌校验策略。
4. 补齐 request-id 与 access log 透传说明。
5. 校验 API 版本路由行为。
6. 为核心中间件补中文函数注释与文件注释。
7. 补齐 `/internal/v1/executions/retries/materialize` 转发路由。

#### 已完成（2026-05-05）

- 已统一 gateway 错误响应 helper，保持 `{ "error": "message" }` 契约。
- 已细化 JWT 失败语义：缺少 token、scheme 错误、token 无效、secret 未配置。
- 已细化 internal token 失败语义：令牌未配置、令牌错误。
- 已标准化 `X-Request-ID` 默认头，并支持配置化 request-id 信任与透传。
- 已补齐 access log 关键字段：method、path、status、duration、ip、request-id、user-agent。
- 已支持 `GATEWAY_API_SUPPORTED_VERSIONS` 配置，并为未知版本返回统一 404 错误。
- 已补齐核心中间件、版本路由、proxy 的中文注释。
- 已为 gateway 增加路由、鉴权、request-id、版本路由相关测试。

#### 验收

- 网关错误边界稳定。
- 中英文前端切换不影响鉴权和转发。
- 网关核心职责清晰。

#### 并行建议

建议尽早启动，优先于大多数业务 worktree 合并。

---

### 5.11 `worktree-iam-service`

#### 目标

完善认证相关实现细节与注释。

#### 任务

1. 登录 / 注册错误映射整理。
2. 用户状态和密码规则提示整理。
3. 认证核心逻辑补中文注释。
4. 必要接口说明和测试补全。

#### 验收

- 认证错误稳定且可解释。
- 核心认证逻辑具备清晰中文说明。

---

### 5.12 `worktree-project-service`

#### 目标

确保项目管理服务与前端双语展示保持契约一致。

#### 任务

1. 项目 CRUD 接口行为检查。
2. 分页、空列表、错误响应整理。
3. 项目核心函数补中文注释。
4. 项目服务文件职责说明补齐。

#### 验收

- 项目列表与创建链路稳定。
- 服务内部逻辑可读性提升。

---

### 5.13 `worktree-spider-service`

#### 目标

完善 Spider、Spider Version 和 registryAuthRef 的接口与逻辑。

#### 任务

1. Spider CRUD 与版本管理接口整理。
2. `registryAuthRef` 传递与校验。
3. 版本唯一性与分页行为整理。
4. 核心版本逻辑补中文注释。
5. 文件职责说明补齐。

#### 验收

- Spider / version 相关接口与设计一致。
- 凭据引用逻辑清晰可追踪。

---

### 5.14 `worktree-execution-service`

#### 目标

完善执行主链路的创建、领取、启动、日志、完成、失败、重试能力。

#### 任务

1. 执行创建与状态流转整理。
2. 资源限制字段行为整理。
3. 重试策略字段与物化逻辑整理。
4. 日志写入与查询接口整理。
5. `registryAuthRef` 继承逻辑整理。
6. 队列、状态机、日志相关核心函数补中文注释。
7. 核心模块文件注释补齐。

#### 验收

- 执行主链路契约稳定。
- 重试和日志行为与设计文档一致。
- 状态机逻辑可读。

---

### 5.15 `worktree-scheduler-service`

#### 目标

完善调度物化、重试驱动和 cron 触发逻辑。

#### 任务

1. Schedule CRUD 行为整理。
2. `last_materialized_at` 游标逻辑整理。
3. catch-up 和防重复物化整理。
4. 重试物化驱动整理。
5. 调度核心函数补中文注释。
6. 调度模块文件注释补齐。

#### 验收

- 调度不会重复物化。
- 调度与重试驱动逻辑清晰。

---

### 5.16 `worktree-node-service`

#### 目标

完善节点心跳、在线态、历史会话和执行过滤能力。

#### 任务

1. 心跳接口与在线状态整理。
2. 节点详情与会话 API 整理。
3. 执行历史过滤参数整理。
4. 节点状态计算逻辑补中文注释。
5. 文件职责说明补齐。

#### 验收

- 节点在线 / 离线判断稳定。
- 历史会话与执行过滤可正确展示。

---

### 5.17 `worktree-datasource-service`

#### 目标

完善数据源配置、连接测试和预览能力。

#### 任务

1. datasource CRUD 行为整理。
2. test / preview 真实探测链路整理。
3. 失败原因映射整理。
4. 数据源连接逻辑补中文注释。
5. 文件注释补齐。

#### 验收

- test / preview 行为稳定且可解释。
- 前端能拿到可翻译的错误语义。

---

### 5.18 `worktree-monitor-service`

#### 目标

完善监控总览、告警规则、告警事件、Webhook 记录能力。

#### 任务

1. 监控总览聚合逻辑整理。
2. 告警规则 CRUD 行为整理。
3. 告警事件持久化与分页查询整理。
4. 告警匹配和发送流程补中文注释。
5. 文件注释补齐。

#### 验收

- 告警规则与事件链路清晰。
- 监控总览数据可重复验证。

---

### 5.19 `worktree-docs-product`

#### 目标

把产品类文档整理成双语可读版本。

#### 任务

1. 产品概览双语化。
2. MVP / smoke checklist 双语化。
3. 功能说明双语化。
4. 使用说明术语统一。

#### 验收

- 中英文用户都能独立阅读产品说明。

---

### 5.20 `worktree-docs-design`

#### 目标

把设计类文档按统一模板双语化。

#### 任务

1. API 设计双语化。
2. 数据库设计双语化。
3. 前端设计双语化。
4. Service Communication 文档双语化。

#### 验收

- 设计文档两种语言下结构一致。
- 术语表被正确引用。

---

### 5.21 `worktree-docs-architecture`

#### 目标

把架构类文档统一成中英双语可读形式。

#### 任务

1. 架构概览双语化。
2. 平台分阶段说明双语化。
3. 服务依赖与通信描述双语化。

#### 验收

- 架构文档可作为跨团队沟通依据。

---

### 5.22 `worktree-docs-howto`

#### 目标

把开发、部署、调试、运维类文档统一双语化。

#### 任务

1. 本地开发文档双语化。
2. Docker Compose 工作流双语化。
3. smoke / 验证流程双语化。
4. 常见问题双语化。

#### 验收

- 新成员可用中文或英文文档完成上手。

---

### 5.23 `worktree-ci-lint`

#### 目标

把注释规范、文档一致性、基础格式检查纳入自动化。

#### 任务

1. 增加中文函数注释检查约定。
2. 增加中文文件注释检查约定。
3. 增加双语文档一致性检查约定。
4. 增加术语表引用检查约定。

#### 验收

- 后续新增内容不容易偏离规范。
- CI 可以提示明显遗漏。

---

### 5.24 `worktree-e2e-smoke`

#### 目标

建立最终回归与冒烟检查，覆盖前端双语与核心业务链路。

#### 任务

1. 登录 / 项目 / Spider / Execution / Schedule / Node / Datasource / Monitor 核心流程 smoke。
2. 中英文切换 smoke。
3. 典型错误状态 smoke。
4. 文档链接与关键术语 smoke。

#### 验收

- 路线图各阶段产物可以统一验收。

---

## 6. 最大并行启动方案

如果目标是最大化并行，推荐采用“公共规范先开，但不等待全部完成才进入下一批”的方式，将 worktree 按边界尽可能一次性铺开。这个方案的核心是：

- 先冻结最容易引发全局返工的契约：术语表、双语文档模板、注释规范、测试模板。
- 然后前端底座、后端服务、文档与验收工作同时分流。
- 每个 worktree 尽量只覆盖一个仓库或一个清晰子域，减少文件交叉修改。

### 6.1 批次 1：先冻结公共规范

这一批是所有后续 worktree 的基础，建议最先启动：

- `worktree-platform-spec`
- `worktree-docs-bilingual-core`
- `worktree-test-harness`

### 6.2 批次 2：前端底座 + 后端服务一次性铺开

这一批尽量同时开启，确保多数仓库可以立刻并行推进：

- `worktree-web-i18n-core`
- `worktree-web-shared-components`
- `worktree-gateway`
- `worktree-iam-service`
- `worktree-project-service`
- `worktree-spider-service`
- `worktree-execution-service`
- `worktree-scheduler-service`
- `worktree-node-service`
- `worktree-datasource-service`
- `worktree-monitor-service`

### 6.3 批次 3：前端业务页面并行

前端底座稳定后，业务页面可以并行进入：

- `worktree-web-auth-project`
- `worktree-web-spider-execution`
- `worktree-web-schedule-node`
- `worktree-web-datasource-monitor`

### 6.4 批次 4：文档与验收收口并行

当功能实现进入稳定期后，文档与质量收口工作可以独立并行：

- `worktree-docs-product`
- `worktree-docs-design`
- `worktree-docs-architecture`
- `worktree-docs-howto`
- `worktree-ci-lint`
- `worktree-e2e-smoke`

### 6.5 最大并行时的推荐组合

如果团队人手充足，可以按下面四组同时推进：

1. 规范组
   - `worktree-platform-spec`
   - `worktree-docs-bilingual-core`
   - `worktree-test-harness`
2. 前端底座组
   - `worktree-web-i18n-core`
   - `worktree-web-shared-components`
3. 后端服务组
   - `worktree-gateway`
   - `worktree-iam-service`
   - `worktree-project-service`
   - `worktree-spider-service`
   - `worktree-execution-service`
   - `worktree-scheduler-service`
   - `worktree-node-service`
   - `worktree-datasource-service`
   - `worktree-monitor-service`
4. 页面与文档组
   - `worktree-web-auth-project`
   - `worktree-web-spider-execution`
   - `worktree-web-schedule-node`
   - `worktree-web-datasource-monitor`
   - `worktree-docs-product`
   - `worktree-docs-design`
   - `worktree-docs-architecture`
   - `worktree-docs-howto`
   - `worktree-ci-lint`
   - `worktree-e2e-smoke`

---

## 7. 依赖关系说明

### 7.1 先决依赖

- `worktree-platform-spec` 是所有 worktree 的公共前提。
- `worktree-docs-bilingual-core` 是所有文档 worktree 的公共前提。
- `worktree-test-harness` 是所有验收 worktree 的公共前提。
- `worktree-web-i18n-core` 是所有前端业务 worktree 的公共前提。

### 7.2 服务依赖

- `worktree-gateway` 建议优先完成，因为它影响全局错误边界。
- `worktree-execution-service`、`worktree-scheduler-service`、`worktree-node-service` 存在链路关系，但可在契约稳定后并行推进。
- `worktree-monitor-service` 依赖执行与节点契约稳定后推进更稳妥。

### 7.3 前端依赖

- 通用组件 worktree 应尽量早于业务页面 worktree 合并。
- 业务页面 worktree 依赖 i18n 底座和相关后端契约。
- 英文文案长度适配应先在通用组件层解决。

### 7.4 文档依赖

- 产品 / 设计 / 架构 / How-to 文档都应优先引用统一术语表。
- 文档 worktree 尽量在功能 worktree 进入稳定期后启动，避免反复改术语。

---

## 8. 风险控制

### 8.1 风险：worktree 修改同一文件导致冲突

**对策：**

- 每个 worktree 约束在明确目录边界内。
- 公共规范先冻结，减少后续分散修改。
- 合并前做文件级冲突检查。

### 8.2 风险：术语和翻译不一致

**对策：**

- 术语表先行。
- 共享 key 与共享文档模板统一维护。
- 文档和前端文案使用同一词库。

### 8.3 风险：英文文案导致布局变化

**对策：**

- 先在通用组件层支持长文案。
- 表格、按钮、弹窗提前预留空间。
- 业务页面切换前先完成公共布局校正。

### 8.4 风险：注释过多或过少

**对策：**

- 只对核心函数和关键文件要求中文注释。
- 明确“不写无意义注释”。
- CI 中加入抽检项。

### 8.5 风险：文档双语维护成本高

**对策：**

- 使用成对文件和共享模板。
- 先做核心文档，再扩展次级文档。
- 明确哪些内容必须全文双语，哪些内容可先摘要后补全。

---

## 9. 最终交付标准

满足以下条件时，可视为本路线图完成：

1. 每个核心仓库都有明确的 worktree 边界。
2. 前端具备稳定的中英双语切换能力。
3. 核心文档具备中英双语版本或双语主体。
4. 新增和重点修改的代码具备中文函数注释与文件注释。
5. 核心服务契约与设计文档一致。
6. smoke / 回归 / 文档一致性检查可以重复执行。
7. 后续新增功能可以直接沿用本路线图中的规范与 worktree 模式。

---

## 10. 建议使用方式

后续如果要启动一个新 worktree，建议按以下顺序：

1. 确认它属于哪一个 worktree 组。
2. 确认依赖的上游 worktree 是否已冻结或合并。
3. 约束该 worktree 只修改所属目录范围内的文件。
4. 先完成最小可验收目标，再扩展次要内容。
5. 合并前按统一 smoke checklist 验收。

这样可以最大限度保留并行开发效率，减少跨 worktree 互相等待。
