# 前端国际化规范

> 适用范围：Web 前端页面、通用组件、表单校验、错误提示和状态展示。

## 语言与资源目录

默认支持：

- `zh-CN`：简体中文
- `en-US`：英文

推荐资源结构：

```text
src/locales/
  zh-CN/
    common.ts
    navigation.ts
    pages.ts
    errors.ts
  en-US/
    common.ts
    navigation.ts
    pages.ts
    errors.ts
```

## Key 命名规则

使用点分层级，按业务语义组织：

```text
common.actions.create
common.actions.delete
common.status.loading
navigation.projects
pages.projects.title
pages.executions.fields.status
errors.gateway.missingBearerToken
```

规则：

1. key 使用英文、小写驼峰或小写单词。
2. 不把完整中文句子当 key。
3. 通用文案放 `common`，页面专有文案放 `pages.<domain>`。
4. 错误消息映射放 `errors`。
5. 后端稳定 `error` 字符串必须有可映射 key。

## 页面标题与导航

页面标题、菜单和面包屑必须来自 i18n 资源，不在组件中硬编码。

示例分类：

| 场景 | key 前缀 |
| --- | --- |
| 主导航 | `navigation.*` |
| 页面标题 | `pages.<domain>.title` |
| 页面描述 | `pages.<domain>.description` |
| 表格列名 | `pages.<domain>.columns.*` |
| 表单字段 | `pages.<domain>.fields.*` |

## 按钮与操作

通用按钮优先复用：

- `common.actions.create`
- `common.actions.edit`
- `common.actions.delete`
- `common.actions.refresh`
- `common.actions.retry`
- `common.actions.confirm`
- `common.actions.cancel`

只有业务语义明显不同的操作才放页面命名空间。

## 状态与枚举

执行、节点、告警等状态展示应集中映射：

```text
common.status.pending
common.status.running
common.status.succeeded
common.status.failed
common.status.online
common.status.offline
```

不要在每个页面重复写状态翻译。

## Fallback 规则

1. 缺失 key 时页面不得崩溃。
2. 缺失翻译可回退到 `zh-CN` 或 key 本身。
3. 开发环境应在控制台提示缺失 key。
4. 生产环境应避免把堆栈或内部错误暴露给用户。

## 语言切换

语言切换应满足：

- 保持当前路由；
- 保持用户已输入但未提交的表单内容；
- 持久化用户选择；
- 刷新页面后继续使用上次语言；
- 切换后重新计算页面标题、菜单、表格列和提示文案。

## 错误提示映射

后端返回：

```json
{ "error": "missing bearer token" }
```

前端应映射到类似：

```text
errors.gateway.missingBearerToken
```

未映射错误使用通用 fallback：

- 中文：`操作失败：{message}`
- 英文：`Operation failed: {message}`

## 英文长度适配

英文文案通常更长，组件需要注意：

1. 按钮允许最小宽度自适应。
2. 表格列支持省略与 tooltip。
3. 弹窗宽度不依赖中文长度。
4. 空状态和错误提示允许换行。
5. 菜单和面包屑避免固定过窄宽度。
