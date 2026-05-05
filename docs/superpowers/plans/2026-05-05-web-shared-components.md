# Web Shared Components (Element Plus) 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Web 前端迁移到 Element Plus，创建 6 个共享组件，迁移 4 个现有组件，补全 i18n key，统一视图风格。

**Architecture:** 所有组件通过 `useLocaleStore().t()` 读取文案，Element Plus 语言包随 locale 响应式切换。工具函数（AppNotification、AppConfirmDialog）直接调用 `ElMessage`/`ElMessageBox`，不依赖 Vue 组件树。

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, Element Plus, @element-plus/icons-vue, Vitest

---

### Task 1: 安装 Element Plus 依赖

**Files:**
- Modify: `apps/web/package.json`

- [ ] **Step 1: 添加 element-plus 和图标包**

```bash
cd apps/web && npm install element-plus @element-plus/icons-vue
```

- [ ] **Step 2: 验证安装成功**

```bash
node -e "require('element-plus/package.json').version" && echo "EP OK"
node -e "require('@element-plus/icons-vue/package.json').version" && echo "Icons OK"
```

Expected: 两个 "OK" 输出，无报错。

- [ ] **Step 3: Commit**

```bash
git add apps/web/package.json apps/web/package-lock.json
git commit -m "chore: add element-plus and @element-plus/icons-vue dependencies"
```

---

### Task 2: 注册 Element Plus（基础注册，语言包在 Task 13 由 App.vue 接管）

**Files:**
- Modify: `apps/web/src/main.ts`

- [ ] **Step 1: 更新 main.ts（最小变更：只加 EP 和 CSS）**

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

createApp(App).use(createPinia()).use(router).use(ElementPlus).mount('#app')
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

Expected: 构建成功，无类型错误。

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/main.ts
git commit -m "feat: register Element Plus in main.ts"
```

---

### Task 3: 迁移 AppLanguageSwitcher → el-select

**Files:**
- Modify: `apps/web/src/components/AppLanguageSwitcher.vue`

- [ ] **Step 1: 重写组件**

```vue
<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLocaleStore } from '../stores/locale'
import type { Locale } from '../i18n/messages'

const localeStore = useLocaleStore()
const { availableLocales, locale } = storeToRefs(localeStore)

const options = [
  { value: 'zh-CN', label: '中文' },
  { value: 'en-US', label: 'English' },
]
</script>

<template>
  <div class="language-switcher">
    <span class="label">{{ localeStore.t('app.language') }}</span>
    <el-select
      :model-value="locale"
      size="small"
      style="width: 100px"
      @update:model-value="localeStore.setLocale($event as Locale)"
    >
      <el-option
        v-for="item in options"
        :key="item.value"
        :label="item.label"
        :value="item.value"
      />
    </el-select>
  </div>
</template>

<style scoped>
.language-switcher {
  align-items: center;
  display: inline-flex;
  gap: 0.5rem;
}
.label {
  white-space: nowrap;
}
</style>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppLanguageSwitcher.vue
git commit -m "feat: migrate AppLanguageSwitcher to el-select"
```

---

### Task 4: 迁移 AppEmptyState → el-empty

**Files:**
- Modify: `apps/web/src/components/AppEmptyState.vue`

- [ ] **Step 1: 重写组件**

```vue
<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    messageKey?: string
  }>(),
  {
    messageKey: 'common.state.empty',
  },
)

const localeStore = useLocaleStore()
</script>

<template>
  <el-empty :description="localeStore.t(props.messageKey)" />
</template>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppEmptyState.vue
git commit -m "feat: migrate AppEmptyState to el-empty"
```

---

### Task 5: 迁移 AppLoadingState → v-loading + el-skeleton

**Files:**
- Modify: `apps/web/src/components/AppLoadingState.vue`

- [ ] **Step 1: 重写组件，支持两种模式**

```vue
<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    mode?: 'skeleton' | 'text'
    loading?: boolean
    messageKey?: string
  }>(),
  {
    mode: 'skeleton',
    loading: true,
    messageKey: 'common.state.loading',
  },
)

const localeStore = useLocaleStore()
</script>

<template>
  <div v-if="props.loading" v-loading="props.mode === 'text'" class="loading-state">
    <template v-if="props.mode === 'skeleton'">
      <el-skeleton :rows="3" animated />
    </template>
    <template v-else>
      <span>{{ localeStore.t(props.messageKey) }}</span>
    </template>
  </div>
</template>

<style scoped>
.loading-state {
  padding: 1rem;
}
</style>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppLoadingState.vue
git commit -m "feat: migrate AppLoadingState to Element Plus skeleton/loading"
```

---

### Task 6: 迁移 AppLayout → el-container/el-menu

**Files:**
- Modify: `apps/web/src/components/AppLayout.vue`

- [ ] **Step 1: 重写组件**

```vue
<script setup lang="ts">
import AppLanguageSwitcher from './AppLanguageSwitcher.vue'
import { useLocaleStore } from '../stores/locale'
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const localeStore = useLocaleStore()
const route = useRoute()

const activeIndex = computed(() => {
  const path = route.path
  if (path.startsWith('/login')) return '/login'
  if (path.startsWith('/projects')) return '/projects'
  if (path.startsWith('/spiders')) return '/spiders'
  if (path.startsWith('/executions')) return '/executions'
  if (path.startsWith('/monitor')) return '/monitor'
  if (path.startsWith('/datasources')) return '/datasources'
  return '/projects'
})
</script>

<template>
  <el-container class="app-layout">
    <el-header class="app-header" height="auto">
      <div class="header-left">
        <router-link class="brand" to="/projects">{{ localeStore.t('app.title') }}</router-link>
        <el-menu
          :default-active="activeIndex"
          mode="horizontal"
          router
          :ellipsis="false"
          class="main-menu"
        >
          <el-menu-item index="/login">{{ localeStore.t('navigation.login') }}</el-menu-item>
          <el-menu-item index="/projects">{{ localeStore.t('navigation.projects') }}</el-menu-item>
          <el-menu-item index="/spiders">{{ localeStore.t('navigation.spiders') }}</el-menu-item>
          <el-menu-item index="/executions">{{ localeStore.t('navigation.executions') }}</el-menu-item>
          <el-menu-item index="/monitor">{{ localeStore.t('navigation.monitor') }}</el-menu-item>
          <el-menu-item index="/datasources">{{ localeStore.t('navigation.datasources') }}</el-menu-item>
        </el-menu>
      </div>
      <AppLanguageSwitcher />
    </el-header>
    <el-main class="app-content">
      <slot />
    </el-main>
  </el-container>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.app-header {
  align-items: center;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  justify-content: space-between;
  padding: 0 1rem;
}

.header-left {
  align-items: center;
  display: flex;
  gap: 1rem;
}

.brand {
  color: var(--el-text-color-primary);
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
}

.main-menu {
  border-bottom: none;
}

.app-content {
  padding: 1rem;
}
</style>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppLayout.vue
git commit -m "feat: migrate AppLayout to el-container/el-menu"
```

---

### Task 7: 创建 AppNotification 工具函数

**Files:**
- Create: `apps/web/src/components/AppNotification.ts`

- [ ] **Step 1: 创建文件**

```typescript
import { ElMessage, ElNotification } from 'element-plus'

type NotificationOptions = {
  messageKey?: string
  message?: string
  titleKey?: string
  title?: string
  duration?: number
}

export function notifySuccess(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.success({
    message: opts.message ?? '',
    duration: opts.duration ?? 3000,
  })
}

export function notifyError(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.error({
    message: opts.message ?? '',
    duration: opts.duration ?? 5000,
  })
}

export function notifyWarning(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.warning({
    message: opts.message ?? '',
    duration: opts.duration ?? 4000,
  })
}

export function notifyInfo(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.info({
    message: opts.message ?? '',
    duration: opts.duration ?? 3000,
  })
}

export function notifySuccessPersistent(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElNotification.success({
    title: opts.title ?? '',
    message: opts.message ?? '',
    duration: opts.duration ?? 0,
  })
}

export function notifyErrorPersistent(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElNotification.error({
    title: opts.title ?? '',
    message: opts.message ?? '',
    duration: opts.duration ?? 0,
  })
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppNotification.ts
git commit -m "feat: add AppNotification utility (ElMessage/ElNotification wrappers)"
```

---

### Task 8: 创建 AppConfirmDialog 工具函数

**Files:**
- Create: `apps/web/src/components/AppConfirmDialog.ts`

- [ ] **Step 1: 创建文件**

```typescript
import { ElMessageBox } from 'element-plus'

type ConfirmOptions = {
  titleKey?: string
  title?: string
  messageKey?: string
  message?: string
  confirmButtonTextKey?: string
  confirmButtonText?: string
  cancelButtonTextKey?: string
  cancelButtonText?: string
  type?: 'warning' | 'info' | 'error'
}

export async function confirmAction(options: ConfirmOptions): Promise<boolean> {
  try {
    await ElMessageBox.confirm(
      options.message ?? options.messageKey ?? 'Are you sure?',
      options.title ?? options.titleKey ?? 'Confirm',
      {
        confirmButtonText: options.confirmButtonText ?? options.confirmButtonTextKey ?? 'OK',
        cancelButtonText: options.cancelButtonText ?? options.cancelButtonTextKey ?? 'Cancel',
        type: options.type ?? 'warning',
      },
    )
    return true
  } catch {
    return false
  }
}

export async function confirmDelete(
  messageKey?: string,
  message?: string,
): Promise<boolean> {
  return confirmAction({
    titleKey: 'common.actions.delete',
    messageKey: messageKey ?? 'common.actions.confirmDelete',
    message,
    confirmButtonTextKey: 'common.actions.confirm',
    cancelButtonTextKey: 'common.actions.cancel',
    type: 'error',
  })
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppConfirmDialog.ts
git commit -m "feat: add AppConfirmDialog utility (ElMessageBox wrapper)"
```

---

### Task 9: 创建 AppErrorState 组件

**Files:**
- Create: `apps/web/src/components/AppErrorState.vue`

- [ ] **Step 1: 创建文件**

```vue
<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    messageKey?: string
    message?: string
    retryTextKey?: string
    retryText?: string
    showRetry?: boolean
  }>(),
  {
    messageKey: 'common.error.default',
    showRetry: true,
  },
)

const emit = defineEmits<{
  retry: []
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <el-result icon="error" :title="props.message ?? localeStore.t(props.messageKey!)">
    <template v-if="props.showRetry" #extra>
      <el-button type="primary" @click="emit('retry')">
        {{ props.retryText ?? props.retryTextKey ? localeStore.t(props.retryTextKey!) : localeStore.t('common.actions.retry') }}
      </el-button>
    </template>
  </el-result>
</template>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppErrorState.vue
git commit -m "feat: add AppErrorState component (el-result wrapper)"
```

---

### Task 10: 创建 AppBreadcrumb 组件

**Files:**
- Create: `apps/web/src/components/AppBreadcrumb.vue`

- [ ] **Step 1: 创建文件**

```vue
<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

export interface BreadcrumbItem {
  to?: string
  labelKey: string
  label?: string
}

defineProps<{
  items: BreadcrumbItem[]
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item v-for="(item, index) in items" :key="index" :to="item.to">
      {{ item.label ?? localeStore.t(item.labelKey) }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppBreadcrumb.vue
git commit -m "feat: add AppBreadcrumb component (el-breadcrumb wrapper)"
```

---

### Task 11: 创建 AppTable 组件

**Files:**
- Create: `apps/web/src/components/AppTable.vue`

- [ ] **Step 1: 创建文件**

```vue
<script setup lang="ts" generic="T extends Record<string, unknown>">
import { useLocaleStore } from '../stores/locale'

export interface AppTableColumn {
  prop: string
  labelKey: string
  label?: string
  width?: string | number
  minWidth?: string | number
  sortable?: boolean | 'custom'
  align?: 'left' | 'center' | 'right'
  formatter?: (row: T, column: AppTableColumn, cellValue: unknown, index: number) => string
}

const props = withDefaults(
  defineProps<{
    columns: AppTableColumn[]
    data: T[]
    loading?: boolean
    total?: number
    page?: number
    pageSize?: number
    pageSizes?: number[]
    showPagination?: boolean
    emptyMessageKey?: string
  }>(),
  {
    loading: false,
    total: 0,
    page: 1,
    pageSize: 10,
    pageSizes: () => [10, 20, 50, 100],
    showPagination: true,
    emptyMessageKey: 'common.state.empty',
  },
)

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [size: number]
  'sort-change': [sort: { prop: string; order: string | null }]
  'row-click': [row: T, column: unknown, event: MouseEvent]
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <div class="app-table">
    <el-table
      :data="props.data"
      v-loading="props.loading"
      stripe
      border
      style="width: 100%"
      @sort-change="emit('sort-change', $event)"
      @row-click="(row, column, event) => emit('row-click', row as T, column, event)"
    >
      <template #empty>
        <el-empty :description="localeStore.t(props.emptyMessageKey!)" />
      </template>
      <el-table-column
        v-for="col in props.columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label ?? localeStore.t(col.labelKey)"
        :width="col.width"
        :min-width="col.minWidth"
        :sortable="col.sortable"
        :align="col.align ?? 'left'"
        :formatter="col.formatter as (row: unknown, column: unknown, cellValue: unknown, index: number) => string"
        show-overflow-tooltip
      />
    </el-table>
    <div v-if="props.showPagination && props.total > 0" class="pagination-wrap">
      <el-pagination
        :current-page="props.page"
        :page-size="props.pageSize"
        :page-sizes="props.pageSizes"
        :total="props.total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @current-change="(p: number) => emit('update:page', p)"
        @size-change="(s: number) => emit('update:pageSize', s)"
      />
    </div>
  </div>
</template>

<style scoped>
.app-table {
  display: grid;
  gap: 1rem;
}
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
}
</style>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppTable.vue
git commit -m "feat: add AppTable component (el-table + el-pagination wrapper)"
```

---

### Task 12: 创建 AppForm 组件

**Files:**
- Create: `apps/web/src/components/AppForm.vue`

- [ ] **Step 1: 创建文件**

```vue
<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'
import type { FormItemRule } from 'element-plus'

export interface AppFormField {
  prop: string
  labelKey: string
  label?: string
  placeholderKey?: string
  placeholder?: string
  type?: 'input' | 'textarea' | 'number' | 'select' | 'switch' | 'date'
  required?: boolean
  rules?: FormItemRule[]
  options?: { label: string; value: string | number }[]
  disabled?: boolean
  rows?: number
}

const props = withDefaults(
  defineProps<{
    fields: AppFormField[]
    modelValue: Record<string, unknown>
    loading?: boolean
    submitTextKey?: string
    cancelTextKey?: string
    showCancel?: boolean
    labelWidth?: string
  }>(),
  {
    loading: false,
    submitTextKey: 'common.actions.confirm',
    cancelTextKey: 'common.actions.cancel',
    showCancel: false,
    labelWidth: '120px',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, unknown>]
  submit: []
  cancel: []
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <el-form
    :model="props.modelValue"
    :label-width="props.labelWidth"
    @submit.prevent="emit('submit')"
  >
    <el-form-item
      v-for="field in props.fields"
      :key="field.prop"
      :prop="field.prop"
      :label="field.label ?? localeStore.t(field.labelKey)"
      :required="field.required"
      :rules="field.rules"
    >
      <el-input
        v-if="!field.type || field.type === 'input'"
        :model-value="props.modelValue[field.prop]"
        :placeholder="field.placeholder ?? (field.placeholderKey ? localeStore.t(field.placeholderKey) : '')"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-input
        v-else-if="field.type === 'textarea'"
        type="textarea"
        :rows="field.rows ?? 3"
        :model-value="props.modelValue[field.prop]"
        :placeholder="field.placeholder ?? (field.placeholderKey ? localeStore.t(field.placeholderKey) : '')"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-input-number
        v-else-if="field.type === 'number'"
        :model-value="props.modelValue[field.prop] as number"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-select
        v-else-if="field.type === 'select'"
        :model-value="props.modelValue[field.prop]"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      >
        <el-option
          v-for="opt in field.options"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-switch
        v-else-if="field.type === 'switch'"
        :model-value="props.modelValue[field.prop] as boolean"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-date-picker
        v-else-if="field.type === 'date'"
        :model-value="props.modelValue[field.prop]"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" :loading="props.loading" native-type="submit">
        {{ localeStore.t(props.submitTextKey!) }}
      </el-button>
      <el-button v-if="props.showCancel" @click="emit('cancel')">
        {{ localeStore.t(props.cancelTextKey!) }}
      </el-button>
    </el-form-item>
  </el-form>
</template>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/components/AppForm.vue
git commit -m "feat: add AppForm component (el-form wrapper)"
```

---

### Task 13: 更新 App.vue 使用 AppLayout

**Files:**
- Modify: `apps/web/src/App.vue`

- [ ] **Step 1: 重写 App.vue**

```vue
<script setup lang="ts">
import { computed, provide } from 'vue'
import { useLocaleStore } from './stores/locale'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import AppLayout from './components/AppLayout.vue'

const localeStore = useLocaleStore()
const elLocale = computed(() => (localeStore.locale === 'zh-CN' ? zhCn : en))
provide('elLocale', elLocale)
</script>

<template>
  <el-config-provider :locale="elLocale">
    <AppLayout>
      <router-view />
    </AppLayout>
  </el-config-provider>
</template>
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/App.vue
git commit -m "feat: wrap app with el-config-provider and AppLayout"
```

---

### Task 14: 补齐 i18n messages 缺失 key

**Files:**
- Modify: `apps/web/src/i18n/messages.ts`

- [ ] **Step 1: 在中文和英文 messages 中补齐以下 key**

中文补充路径（在现有 messages['zh-CN'] 结构中添加）：

```typescript
// common 下补充:
common: {
  // ... 保留现有
  state: {
    // ... 保留现有
    error: '加载失败',
  },
  error: {
    default: '操作失败，请稍后重试',
    network: '网络连接异常',
    unauthorized: '登录已过期，请重新登录',
    forbidden: '无权限访问',
    notFound: '资源不存在',
    serverError: '服务异常',
  },
  actions: {
    // ... 保留现有
    search: '搜索',
    reset: '重置',
    submit: '提交',
    back: '返回',
    save: '保存',
    close: '关闭',
    confirmDelete: '确定要删除吗？此操作不可撤销。',
  },
},

// navigation 下补充:
navigation: {
  // ... 保留现有
  main: '主导航',
  nodes: '节点',
  schedules: '调度',
},

// pages 下补充:
pages: {
  // ... 保留现有 login, projects
  spiders: {
    title: '爬虫管理',
    placeholder: 'Web MVP 外壳占位页。',
  },
  executions: {
    title: '执行管理',
    createTitle: '创建执行',
    detailTitle: '执行详情',
    lookupTitle: '查看执行',
    projectId: '项目 ID',
    spiderId: '爬虫 ID',
    image: '镜像',
    command: '命令',
    placeholder: 'Web MVP 外壳占位页。',
    creating: '创建中...',
    createAction: '创建执行',
    openAction: '打开',
    lookupPlaceholder: '执行 ID',
    createdMessage: '已创建执行',
  },
  monitor: {
    title: '监控',
    placeholder: 'Web MVP 外壳占位页。',
  },
  datasources: {
    title: '数据源',
    placeholder: 'Web MVP 外壳占位页。',
  },
  nodes: {
    title: '节点',
    placeholder: 'Web MVP 外壳占位页。',
  },
  schedules: {
    title: '调度',
    placeholder: 'Web MVP 外壳占位页。',
  },
  notFound: {
    title: '404',
    description: '页面不存在',
  },
},
```

英文补充路径（同样的结构，英文翻译）：

```typescript
'en-US': {
  // ...
  common: {
    // ... 保留现有
    state: {
      // ... 保留现有
      error: 'Load failed',
    },
    error: {
      default: 'Operation failed, please try again later',
      network: 'Network error',
      unauthorized: 'Session expired, please sign in again',
      forbidden: 'Access denied',
      notFound: 'Resource not found',
      serverError: 'Service error',
    },
    actions: {
      // ... 保留现有
      search: 'Search',
      reset: 'Reset',
      submit: 'Submit',
      back: 'Back',
      save: 'Save',
      close: 'Close',
      confirmDelete: 'Are you sure? This action cannot be undone.',
    },
  },
  navigation: {
    // ... 保留现有
    main: 'Main Navigation',
    nodes: 'Nodes',
    schedules: 'Schedules',
  },
  pages: {
    // ... 保留现有 login, projects
    spiders: {
      title: 'Spiders',
      placeholder: 'Web MVP shell placeholder.',
    },
    executions: {
      title: 'Executions',
      createTitle: 'Create Execution',
      detailTitle: 'Execution Detail',
      lookupTitle: 'Open Execution',
      projectId: 'Project ID',
      spiderId: 'Spider ID',
      image: 'Image',
      command: 'Command',
      placeholder: 'Web MVP shell placeholder.',
      creating: 'Creating...',
      createAction: 'Create Execution',
      openAction: 'Open',
      lookupPlaceholder: 'execution id',
      createdMessage: 'Created execution',
    },
    monitor: {
      title: 'Monitor',
      placeholder: 'Web MVP shell placeholder.',
    },
    datasources: {
      title: 'Datasources',
      placeholder: 'Web MVP shell placeholder.',
    },
    nodes: {
      title: 'Nodes',
      placeholder: 'Web MVP shell placeholder.',
    },
    schedules: {
      title: 'Schedules',
      placeholder: 'Web MVP shell placeholder.',
    },
    notFound: {
      title: '404',
      description: 'Page not found',
    },
  },
}
```

- [ ] **Step 2: 验证编译和测试**

```bash
cd apps/web && npx vitest run
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/i18n/messages.ts
git commit -m "feat: add missing i18n keys for all pages and common components"
```

---

### Task 15: 增强 api/client.ts 错误处理

**Files:**
- Modify: `apps/web/src/api/client.ts`

- [ ] **Step 1: 更新 apiFetch 以区分错误类型**

```typescript
const defaultBaseURL = '/api/v1'

function resolveBaseURL() {
  const envValue = import.meta.env.VITE_API_BASE_URL as string | undefined
  return (envValue && envValue.trim()) || defaultBaseURL
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

const defaultErrorCodes: Record<number, string> = {
  400: 'errors.gateway.badRequest',
  401: 'errors.gateway.missingBearerToken',
  403: 'errors.gateway.invalidBearerToken',
  404: 'errors.gateway.notFound',
  429: 'errors.gateway.rateLimitExceeded',
  500: 'errors.gateway.upstreamServiceUnavailable',
  502: 'errors.gateway.upstreamServiceUnavailable',
  503: 'errors.gateway.upstreamServiceUnavailable',
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${resolveBaseURL()}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  if (!response.ok) {
    let body: Record<string, unknown> = {}
    try {
      body = await response.json()
    } catch {
      // 非 JSON 响应体
    }

    const message = (body.message as string) || `request failed: ${response.status}`
    const code = (body.code as string) || (defaultErrorCodes[response.status] ?? 'errors.fallback')

    const err = new ApiError(response.status, code, message)
    throw err
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/api/client.ts
git commit -m "feat: enhance apiFetch with structured error codes and ApiError class"
```

---

### Task 16: 更新所有 views 使用 EP 组件和 i18n

**Files:**
- Modify: `apps/web/src/views/LoginView.vue`
- Modify: `apps/web/src/views/ProjectsView.vue`
- Modify: `apps/web/src/views/SpidersView.vue`
- Modify: `apps/web/src/views/ExecutionsView.vue`
- Modify: `apps/web/src/views/ExecutionDetailView.vue`
- Modify: `apps/web/src/views/SchedulesView.vue`
- Modify: `apps/web/src/views/NodesView.vue`
- Modify: `apps/web/src/views/DatasourcesView.vue`
- Modify: `apps/web/src/views/MonitorView.vue`

- [ ] **Step 1: 更新 ExecutionsView.vue（最复杂的视图）**

```vue
<template>
  <div class="page">
    <el-card>
      <template #header>
        <h1>{{ localeStore.t('pages.executions.title') }}</h1>
      </template>
      <p>{{ localeStore.t('pages.executions.placeholder') }}</p>
    </el-card>

    <el-card>
      <template #header>
        <h2>{{ localeStore.t('pages.executions.createTitle') }}</h2>
      </template>
      <el-form label-width="120px" @submit.prevent="submit">
        <el-form-item :label="localeStore.t('pages.executions.projectId')">
          <el-input v-model="form.projectId" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.spiderId')">
          <el-input v-model="form.spiderId" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.image')">
          <el-input v-model="form.image" placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.command')">
          <el-input v-model="form.command" placeholder="./go-echo" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" native-type="submit">
            {{ submitting ? localeStore.t('pages.executions.creating') : localeStore.t('pages.executions.createAction') }}
          </el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="error" :title="error" type="error" show-icon closable />
      <el-alert v-if="createdExecutionId" type="success" show-icon>
        <template #title>
          {{ localeStore.t('pages.executions.createdMessage') }}
          <router-link :to="`/executions/${createdExecutionId}`">{{ createdExecutionId }}</router-link>
        </template>
      </el-alert>
    </el-card>

    <el-card>
      <template #header>
        <h2>{{ localeStore.t('pages.executions.lookupTitle') }}</h2>
      </template>
      <div class="toolbar">
        <el-input v-model="lookupId" :placeholder="localeStore.t('pages.executions.lookupPlaceholder')" style="max-width: 300px" />
        <el-button @click="openExecution">{{ localeStore.t('pages.executions.openAction') }}</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLocaleStore } from '../stores/locale'
import { createExecution } from '../api/executions'

const router = useRouter()
const localeStore = useLocaleStore()

const form = reactive({
  projectId: 'project-1',
  spiderId: '',
  image: '',
  command: '',
})

const lookupId = ref('')
const createdExecutionId = ref('')
const error = ref('')
const submitting = ref(false)

function parseCommand(input: string) {
  return input
    .split(' ')
    .map((item) => item.trim())
    .filter(Boolean)
}

async function submit() {
  submitting.value = true
  error.value = ''

  try {
    const execution = await createExecution({
      projectId: form.projectId,
      spiderId: form.spiderId,
      image: form.image,
      command: parseCommand(form.command),
    })
    createdExecutionId.value = execution.id
    lookupId.value = execution.id
    await router.push(`/executions/${execution.id}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to create execution'
  } finally {
    submitting.value = false
  }
}

async function openExecution() {
  if (!lookupId.value.trim()) {
    return
  }
  await router.push(`/executions/${lookupId.value.trim()}`)
}
</script>

<style scoped>
.page {
  display: grid;
  gap: 1rem;
}
.toolbar {
  align-items: center;
  display: flex;
  gap: 0.75rem;
}
</style>
```

- [ ] **Step 2: 更新其他视图为统一占位模板**

所有除 Executions 和 ExecutionDetail 外的视图使用统一模板：

```vue
<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
</script>

<template>
  <el-card>
    <template #header>
      <h1>{{ localeStore.t('pages.<page>.title') }}</h1>
    </template>
    <p>{{ localeStore.t('pages.<page>.placeholder') }}</p>
  </el-card>
</template>
```

具体映射：
- LoginView: `pages.login.title` / `pages.login.placeholder`
- ProjectsView: `pages.projects.title` / `pages.projects.placeholder`
- SpidersView: `pages.spiders.title` / `pages.spiders.placeholder`
- SchedulesView: `pages.schedules.title` / `pages.schedules.placeholder`
- NodesView: `pages.nodes.title` / `pages.nodes.placeholder`
- DatasourcesView: `pages.datasources.title` / `pages.datasources.placeholder`
- MonitorView: `pages.monitor.title` / `pages.monitor.placeholder`

ExecutionDetailView 保持现有逻辑，包装在 `el-card` 中。

- [ ] **Step 3: 验证编译通过**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 4: Commit**

```bash
git add apps/web/src/views/
git commit -m "feat: update all views to use Element Plus components and i18n keys"
```

---

### Task 17: 更新测试

**Files:**
- Modify: `apps/web/src/i18n/translate.spec.ts`
- Modify: `apps/web/src/stores/__tests__/locale.spec.ts`
- Modify: `apps/web/src/views/__tests__/executions.spec.ts`

- [ ] **Step 1: 补充 translate.spec.ts 测试新的 i18n key**

```typescript
import { describe, expect, it } from 'vitest'
import { normalizeLocale, translate } from './translate'

describe('translate', () => {
  it('returns translated text for the active locale', () => {
    expect(translate('navigation.login', 'zh-CN')).toBe('登录')
    expect(translate('navigation.login', 'en-US')).toBe('Login')
  })

  it('falls back to the key when the message is missing', () => {
    expect(translate('missing.key', 'zh-CN')).toBe('missing.key')
  })

  it('interpolates placeholder params', () => {
    expect(translate('errors.fallback', 'zh-CN', { message: '未授权' })).toBe('操作失败：未授权')
  })

  it('returns new common action keys', () => {
    expect(translate('common.actions.search', 'zh-CN')).toBe('搜索')
    expect(translate('common.actions.search', 'en-US')).toBe('Search')
  })

  it('returns new error keys', () => {
    expect(translate('common.error.default', 'zh-CN')).toBe('操作失败，请稍后重试')
    expect(translate('common.error.default', 'en-US')).toBe('Operation failed, please try again later')
  })

  it('returns new page keys', () => {
    expect(translate('pages.executions.title', 'zh-CN')).toBe('执行管理')
    expect(translate('pages.executions.title', 'en-US')).toBe('Executions')
    expect(translate('pages.nodes.title', 'en-US')).toBe('Nodes')
    expect(translate('pages.nodes.title', 'zh-CN')).toBe('节点')
  })
})

describe('normalizeLocale', () => {
  it('returns default locale for unsupported values', () => {
    expect(normalizeLocale('fr-FR')).toBe('zh-CN')
    expect(normalizeLocale(null)).toBe('zh-CN')
  })
})
```

- [ ] **Step 2: 运行全部测试**

```bash
cd apps/web && npx vitest run
```

Expected: 所有测试通过。

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/i18n/translate.spec.ts
git commit -m "test: add i18n coverage for new shared component keys"
```

---

### Task 18: 最终验证与修复

**Files:**
- 全部已修改文件

- [ ] **Step 1: TypeScript 类型检查**

```bash
cd apps/web && npx tsc --noEmit
```

- [ ] **Step 2: Vite 构建**

```bash
cd apps/web && npx vite build --logLevel error
```

- [ ] **Step 3: 运行全部测试**

```bash
cd apps/web && npx vitest run
```

- [ ] **Step 4: 如有失败，逐项修复后重新验证**

- [ ] **Step 5: Final commit（如有多轮修复）**

```bash
git add -A && git diff --cached --stat
git commit -m "chore: final validation fixes for web shared components"
```

---

### Task 19: 合并到本地 main 分支

- [ ] **Step 1: 切回主工作树并合并**

```bash
cd /home/iambaby/goland_projects/crawler-platform
git checkout main
git merge worktree-web-shared-components --no-ff -m "$(cat <<'EOF'
feat: add web shared components with Element Plus

- Install Element Plus and @element-plus/icons-vue
- Migrate AppLayout, AppLanguageSwitcher, AppEmptyState, AppLoadingState to EP
- Add AppTable, AppForm, AppConfirmDialog, AppNotification, AppErrorState, AppBreadcrumb
- Wrap app with el-config-provider for dynamic locale switching
- Complete i18n keys for all pages and common components
- Enhance apiFetch with structured ApiError class
- Update all views to use EP components and i18n keys

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 2: 验证合并后状态**

```bash
git log --oneline -5
git status
```

Expected: 干净的工作树，HEAD 在 main 上，无未提交变更。

---

### Task 20: 推送到远端 GitHub

- [ ] **Step 1: 推送 main 分支**

```bash
git push origin main
```

- [ ] **Step 2: 推送 worktree 分支（可选，用于备份）**

```bash
git push origin worktree-web-shared-components
```

---

### Task 21: 更新实现计划文档

**Files:**
- Modify: `docs/product/implementation-plan.md`

- [ ] **Step 1: 在 4.1 Web 前端仓库 下添加共享组件完成记录**

在 `docs/product/implementation-plan.md` 的 `4.1.2 需要实现的内容` 末尾追加完成状态：

```markdown
##### G. 共享组件（已完成）

基于 Element Plus 完成了以下共享组件开发和迁移：

现有组件迁移：
1. AppLayout → el-container + el-header + el-menu（horizontal 模式，路由联动高亮）
2. AppLanguageSwitcher → el-select（小尺寸，中/英文标签）
3. AppEmptyState → el-empty（description 走 i18n）
4. AppLoadingState → el-skeleton + v-loading 指令（skeleton/text 双模式）

新增共享组件：
1. AppTable — 通用表格（el-table + el-pagination），支持排序/分页/加载/空状态，列头走 i18n
2. AppForm — 通用表单（el-form），支持 input/textarea/number/select/switch/date 六种字段类型
3. AppConfirmDialog — 操作确认弹窗（ElMessageBox 封装），含 confirmDelete 便捷函数
4. AppNotification — Toast 通知工具（ElMessage/ElNotification），success/error/warning/info
5. AppErrorState — 错误状态展示 + 重试按钮（el-result）
6. AppBreadcrumb — 面包屑导航（el-breadcrumb）

相关增强：
- el-config-provider 语言包随 localeStore 响应式切换
- apiFetch 增强为 ApiError 类，区分 HTTP 错误码
- messages.ts 补齐 50+ 个 i18n key
- 全部 9 个视图页面迁移到 EP 组件 + i18n
```

- [ ] **Step 2: Commit**

```bash
git add docs/product/implementation-plan.md
git commit -m "docs: mark web shared components tasks as completed"
```

- [ ] **Step 3: 推送到远端**

```bash
git push origin main
```
