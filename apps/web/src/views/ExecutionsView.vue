<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { Execution } from '../api/executions'
import { createExecution, listExecutions } from '../api/executions'
import { ApiError } from '../api/client'
import { useLocaleStore } from '../stores/locale'

const router = useRouter()
const localeStore = useLocaleStore()

const executions = ref<Execution[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const statusFilter = ref('')
const loading = ref(false)
const loadError = ref('')

const showCreateDialog = ref(false)
const createForm = reactive({
  projectId: 'project-1',
  spiderId: '',
  image: '',
  command: '',
})
const creating = ref(false)
const lookupId = ref('')

const statusOptions = [
  { label: 'common.status.pending', value: 'pending' },
  { label: 'common.status.running', value: 'running' },
  { label: 'common.status.succeeded', value: 'succeeded' },
  { label: 'common.status.failed', value: 'failed' },
]

function statusTagType(status: string) {
  const map: Record<string, string> = {
    pending: 'info',
    running: 'warning',
    succeeded: 'success',
    failed: 'danger',
  }
  return map[status] ?? 'info'
}

function statusLabel(status: string) {
  return localeStore.t(`common.status.${status}`) ?? status
}

async function loadExecutions() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listExecutions({
      limit: pageSize.value,
      offset: (page.value - 1) * pageSize.value,
      status: statusFilter.value || undefined,
    })
    executions.value = res.items ?? []
    total.value = res.total ?? 0
  } catch {
    loadError.value = localeStore.t('pages.executions.listLoadFailed')
    executions.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function onPageChange(newPage: number) {
  page.value = newPage
  loadExecutions()
}

function onPageSizeChange(newSize: number) {
  pageSize.value = newSize
  page.value = 1
  loadExecutions()
}

watch(statusFilter, () => {
  page.value = 1
  loadExecutions()
})

function onRowClick(row: Execution) {
  router.push(`/executions/${row.id}`)
}

function openCreateDialog() {
  createForm.projectId = 'project-1'
  createForm.spiderId = ''
  createForm.image = ''
  createForm.command = ''
  showCreateDialog.value = true
}

function parseCommand(input: string) {
  return input
    .split(' ')
    .map((item) => item.trim())
    .filter(Boolean)
}

async function handleCreate() {
  creating.value = true
  try {
    const execution = await createExecution({
      projectId: createForm.projectId,
      spiderId: createForm.spiderId,
      image: createForm.image,
      command: parseCommand(createForm.command),
    })
    ElMessage.success(localeStore.t('pages.executions.createdMessage') + execution.id)
    showCreateDialog.value = false
    await loadExecutions()
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    } else {
      ElMessage.error(localeStore.t('pages.executions.errors.createFailed'))
    }
  } finally {
    creating.value = false
  }
}

function openExecution() {
  if (!lookupId.value.trim()) return
  router.push(`/executions/${lookupId.value.trim()}`)
}

onMounted(loadExecutions)
</script>

<template>
  <main class="executions-page">
    <div class="page-header">
      <h1>{{ localeStore.t('pages.executions.title') }}</h1>
      <el-button type="primary" @click="openCreateDialog">
        {{ localeStore.t('pages.executions.createAction') }}
      </el-button>
    </div>

    <div class="filter-bar">
      <el-select
        v-model="statusFilter"
        :placeholder="localeStore.t('pages.executions.filterStatus')"
        clearable
        style="width: 160px"
      >
        <el-option :label="localeStore.t('pages.executions.allStatus')" value="" />
        <el-option
          v-for="o in statusOptions"
          :key="o.value"
          :label="localeStore.t(o.label)"
          :value="o.value"
        />
      </el-select>
    </div>

    <el-table
      v-loading="loading"
      :data="executions"
      stripe
      :empty-text="localeStore.t('pages.executions.empty')"
      highlight-current-row
      @row-click="onRowClick"
    >
      <el-table-column prop="id" :label="localeStore.t('pages.executions.id')" min-width="200" />
      <el-table-column prop="projectId" :label="localeStore.t('pages.executions.projectId')" min-width="140" />
      <el-table-column prop="spiderId" :label="localeStore.t('pages.executions.spiderId')" min-width="140" />
      <el-table-column :label="localeStore.t('pages.executions.status')" min-width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="triggerSource" :label="localeStore.t('pages.executions.trigger')" min-width="120" />
      <el-table-column prop="createdAt" :label="localeStore.t('pages.executions.createdAt')" min-width="180" />
    </el-table>

    <div v-if="total > 0" class="pagination-wrap">
      <el-pagination
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="onPageChange"
        @size-change="onPageSizeChange"
      />
    </div>

    <el-alert
      v-if="loadError"
      :title="loadError"
      type="error"
      show-icon
      class="error-alert"
    />

    <el-dialog
      v-model="showCreateDialog"
      :title="localeStore.t('pages.executions.createTitle')"
    >
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="localeStore.t('pages.executions.projectId')">
          <el-input v-model="createForm.projectId" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.spiderId')">
          <el-input v-model="createForm.spiderId" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.image')">
          <el-input v-model="createForm.image" placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.command')">
          <el-input v-model="createForm.command" placeholder="./go-echo" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">
          {{ localeStore.t('common.actions.cancel') }}
        </el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">
          {{ localeStore.t('common.actions.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-card class="lookup-card">
      <template #header>
        <h2>{{ localeStore.t('pages.executions.lookupTitle') }}</h2>
      </template>
      <div class="toolbar">
        <el-input
          v-model="lookupId"
          :placeholder="localeStore.t('pages.executions.lookupPlaceholder')"
          style="max-width: 300px"
          @keyup.enter="openExecution"
        />
        <el-button @click="openExecution">
          {{ localeStore.t('pages.executions.openAction') }}
        </el-button>
      </div>
    </el-card>
  </main>
</template>

<style scoped>
.executions-page {
  padding: 0;
}

.page-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.page-header h1 {
  margin: 0;
}

.filter-bar {
  margin-bottom: 1rem;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.error-alert {
  margin-top: 1rem;
}

.lookup-card {
  margin-top: 1rem;
}

.toolbar {
  align-items: center;
  display: flex;
  gap: 0.75rem;
}
</style>
