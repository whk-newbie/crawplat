# 双语文档规范

> 适用范围：产品文档、设计文档、架构文档、开发/部署/运维文档。

## 文件组织

推荐使用成对文件：

```text
docs/design/API_DESIGN.zh-CN.md
docs/design/API_DESIGN.en-US.md
```

如果当前目录已有历史单语文件，可先保留原文件，并在后续文档 worktree 中逐步迁移。

## 语言标识

| 语言 | 后缀 |
| --- | --- |
| 简体中文 | `.zh-CN.md` |
| 英文 | `.en-US.md` |

临时过渡阶段允许：

- 中文主文档：`*.md`
- 英文文档：`*.en-US.md`

但新文档应优先使用显式语言后缀。

## 标题层级同步

双语文档必须保持相同标题层级和顺序。

示例：

```text
## 1. Overview
### 1.1 Scope
### 1.2 Non-goals
```

对应中文：

```text
## 1. 概述
### 1.1 范围
### 1.2 非目标
```

## 表格与接口示例同步

以下内容必须同步：

- API path；
- HTTP method；
- 状态码；
- JSON 字段名；
- 环境变量；
- 配置 key；
- 命令行示例；
- 表格列数量和顺序。

代码块中的字段名不翻译。

## 术语使用

1. 核心术语遵循 `docs/standards/platform-terminology.md`。
2. 首次出现可写“中文（English）”。
3. 后续保持同一种译法。
4. 不为了语言自然度改变 API 字段或状态值。

## 未翻译内容占位

如果必须先提交一侧文档，另一侧使用明确占位：

```markdown
> Translation pending. Source document: `xxx.zh-CN.md`.
```

或：

```markdown
> 待翻译。源文档：`xxx.en-US.md`。
```

禁止使用模糊占位，如 `TODO`、`TBD`，除非同时说明来源和补齐计划。

## 更新流程

修改双语文档时：

1. 先更新源语言文档；
2. 同步更新另一语言标题和结构；
3. 同步表格、接口示例和状态码；
4. 检查术语表；
5. 在 PR / 合并说明中注明是否存在未翻译占位。

## 文档分类建议

| 类型 | 推荐目录 |
| --- | --- |
| API 设计 | `docs/design/` |
| 架构说明 | `docs/architecture/` |
| 产品说明 | `docs/product/` |
| 开发/部署/运维 | `docs/howto/` |
| 规范 | `docs/standards/` |
| 计划与规格 | `docs/superpowers/` |

## 验收清单

- [ ] 文件命名符合语言后缀规则。
- [ ] 标题层级一致。
- [ ] 表格和 API 示例一致。
- [ ] 术语符合平台术语表。
- [ ] 未翻译内容有明确来源和补齐计划。
