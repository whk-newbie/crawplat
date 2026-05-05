# CI Lint Conventions

> This document defines conventions for automated checks on code comments, documentation consistency, and terminology usage. It serves as a reference baseline for CI rule configuration and code review.

## 1. Chinese Comment Checks

### 1.1 Function Comment Requirements

The following types of functions must include Chinese comments:

- Core business functions (main flows: create, query, update, delete)
- Data transformation functions (DTO conversion, serialization, deserialization)
- State calculation functions (state machine transitions, state determination)
- External call wrapper functions (HTTP calls, database operation wrappers)
- Complex conditional branching functions (more than 3 branches)
- Retry / scheduling / message delivery / sync logic functions

Comments should explain:

- What the function does
- Input and output semantics
- Key side effects or boundary conditions
- Non-obvious implementation rationale

### 1.2 File Comment Requirements

The following types of files must include Chinese file comments:

- Business modules with a clear single responsibility
- Adaptation layers interacting with external systems
- Complex state machines / queues / scheduling / alert processing modules
- Shared utility classes, shared configuration, and shared constant files

File comments should explain:

- What the file is responsible for
- Who it interacts with
- What it is NOT responsible for

### 1.3 Comment Quality Rules

- Do not add pointless comments to obvious single-line code
- Do not write "translation comments" that merely repeat what the code says
- Prioritize adding explanations for complex logic, public boundaries, and easily misunderstood areas
- Comments must be updated in sync with implementation changes; avoid stale comments

### 1.4 Verification Methods

Recommended CI verification approaches:

- Use `grep` or custom scripts to check if new Go files contain Chinese comment patterns `// .*[\x{4e00}-\x{9fff}]`
- Sample-check key module directories (`apps/*/internal/service/`, `apps/*/internal/model/`)
- Human review of comment quality during PR review

## 2. Bilingual Documentation Consistency Checks

### 2.1 File Pairing Check

For bilingual documents in the `docs/` directory, check the following rules:

- Each `.zh-CN.md` file should have a corresponding `.en-US.md` file (applies to core docs)
- Each `.en-US.md` file should have a corresponding `.zh-CN.md` file (applies to core docs)
- Example check script:

```bash
# Check for corresponding English docs in the product directory
for f in docs/product/*.zh-CN.md; do
  en="${f/.zh-CN.md/.en-US.md}"
  [ -f "$en" ] || echo "Missing: $en"
done
```

### 2.2 Heading Level Check

- Bilingual documents must maintain the same heading hierarchy (`#`, `##`, `###`, etc.)
- Heading counts should match
- Example check script:

```bash
# Compare heading counts between two documents
zh_headings=$(grep -c '^#' "doc.zh-CN.md")
en_headings=$(grep -c '^#' "doc.en-US.md")
[ "$zh_headings" -eq "$en_headings" ] || echo "Heading count mismatch"
```

### 2.3 Document Update Check

- When modifying core bilingual documents, PRs must include changes to both language versions
- If full translation is not immediately possible, an explicit marker must be included in the document:
  - `> Translation pending. Source document: xxx.zh-CN.md.`
  - `> 待翻译。源文档：xxx.en-US.md。`
- Ambiguous placeholders such as `TODO` or `TBD` are not allowed

## 3. Terminology Consistency Checks

### 3.1 Terminology Table Reference

Core terminology must use `docs/standards/platform-terminology.md` as the authoritative source.

Key terminology list:

| Chinese | English |
|---------|---------|
| 项目 | Project |
| 爬虫 | Spider |
| 爬虫版本 | Spider Version |
| 执行 | Execution |
| 调度 | Schedule |
| 节点 | Node |
| 数据源 | Datasource |
| 告警规则 | Alert Rule |
| 告警事件 | Alert Event |
| 镜像仓库凭据引用 | Registry Auth Ref |

### 3.2 Terminology Check Rules

- The same concept must use a consistent translation across documentation and code comments
- API field names, status values, and error codes are not translated
- Frontend i18n key text must match the terminology table

## 4. PR Checklist

Each PR should include the following self-check:

- [ ] New or modified core functions have Chinese comments
- [ ] New or modified key files have Chinese file descriptions
- [ ] Bilingual document changes are committed as pairs
- [ ] Terminology usage matches the terminology table
- [ ] No ambiguous `TODO` / `TBD` placeholders
- [ ] Comments are not out of sync with code implementation
