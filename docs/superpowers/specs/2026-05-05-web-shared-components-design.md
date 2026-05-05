# Web Shared Components 设计文档

> 日期: 2026-05-05
> 分支: worktree-web-shared-components
> 状态: 已批准

## 目标

将 Web 前端迁移到 Element Plus UI 库，创建一套可复用的共享组件，统一页面风格和交互模式，同时保证中英双语能力不退化。

## 依赖变更

- 新增 `element-plus`
- 新增 `@element-plus/icons-vue`
- `main.ts` 注册 EP，语言包跟随 `localeStore` 动态切换

## 组件设计

### 现有组件迁移（4个）

| 组件 | EP 方案 |
|------|---------|
| AppLayout | `el-container` + `el-header` + `el-menu`（horizontal），语言切换入口 `el-select` |
| AppLanguageSwitcher | 原生 `<select>` → `el-select` size=small |
| AppEmptyState | 手写虚线框 → `el-empty`，description 走 `localeStore.t()` |
| AppLoadingState | 纯文字 → `el-skeleton` 或 `v-loading` 指令透出 |

### 新增共享组件（6个）

| 组件 | 类型 | 核心 EP 组件 | 接口要点 |
|------|------|-------------|---------|
| AppTable | .vue | `el-table` + `el-pagination` | columns（含 labelKey）、data、loading、total、page/pageSize、@update |
| AppForm | .vue | `el-form` + `el-form-item` | fields（含 labelKey/placeholderKey/rules）、modelValue、@submit |
| AppConfirmDialog | .ts | `ElMessageBox` | 函数式调用，i18n 标题/内容/按钮 |
| AppNotification | .ts | `ElMessage` / `ElNotification` | success/error/warning/info 四个导出函数 |
| AppErrorState | .vue | `el-result` | messageKey、retryTextKey、@retry |
| AppBreadcrumb | .vue | `el-breadcrumb` | items（含 to/labelKey） |

### i18n 集成原则

- 所有组件 props 使用 `messageKey` / `labelKey` 等后缀区分语言 key 和直接文本
- 组件内部通过 `useLocaleStore().t()` 解析
- Element Plus 自身的语言包（分页、日期选择器等）跟随 `locale` 响应式切换
- 错误消息统一经 `api/client.ts` 拦截，映射为前端可翻译 key

## 文件变更

```
新增:
  apps/web/src/components/AppTable.vue
  apps/web/src/components/AppForm.vue
  apps/web/src/components/AppConfirmDialog.ts
  apps/web/src/components/AppNotification.ts
  apps/web/src/components/AppErrorState.vue
  apps/web/src/components/AppBreadcrumb.vue

修改:
  apps/web/package.json                          (+ element-plus, @element-plus/icons-vue)
  apps/web/src/main.ts                           (+ EP 注册 + 语言包响应式切换)
  apps/web/src/App.vue                           (改用 AppLayout 包裹)
  apps/web/src/components/AppLayout.vue           (EP 重写)
  apps/web/src/components/AppEmptyState.vue       (→ el-empty)
  apps/web/src/components/AppLoadingState.vue     (→ el-skeleton / v-loading)
  apps/web/src/components/AppLanguageSwitcher.vue (→ el-select)
  apps/web/src/i18n/messages.ts                  (补齐 schedule/node/error 相关 key)
  apps/web/src/api/client.ts                     (错误响应解析增强)
  apps/web/src/views/*.vue                       (替换硬编码文案为 i18n key + EP 组件)
  apps/web/index.html                            (lang 属性动态绑定)
```

## 不纳入范围

- 路由守卫 / 权限控制
- 后端服务逻辑修改
- 文档双语化
- CI/CD 配置
