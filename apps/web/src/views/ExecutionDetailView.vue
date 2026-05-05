<template>
  <div class="page">
    <el-card>
      <template #header>
        <h1>{{ localeStore.t('pages.executions.detailTitle') }}</h1>
      </template>
      <AppLoadingState v-if="loading" :rows="5" />
      <AppErrorState v-else-if="error" :message="error" @retry="loadExecution(route.params.id as string)" />
      <el-descriptions v-else-if="execution" :column="2" border>
        <el-descriptions-item :label="localeStore.t('pages.executions.id')">
          {{ execution.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="localeStore.t('pages.executions.status')">
          <el-tag :type="statusTagType(execution.status)" size="small">
            {{ execution.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="localeStore.t('pages.executions.node')">
          {{ execution.nodeId || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="localeStore.t('pages.executions.trigger')">
          {{ execution.triggerSource }}
        </el-descriptions-item>
        <el-descriptions-item :label="localeStore.t('pages.executions.image')">
          {{ execution.image }}
        </el-descriptions-item>
        <el-descriptions-item :label="localeStore.t('pages.executions.command')">
          {{ execution.command?.length ? execution.command.join(' ') : '-' }}
        </el-descriptions-item>
        <el-descriptions-item
          v-if="execution.errorMessage"
          :span="2"
          :label="localeStore.t('pages.executions.error')"
        >
          <span class="error-text">{{ execution.errorMessage }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card>
      <template #header>
        <h2>{{ localeStore.t('pages.executions.logs') }}</h2>
      </template>
      <el-timeline v-if="logs.length">
        <el-timeline-item
          v-for="entry in logs"
          :key="entry.id"
          :timestamp="entry.createdAt"
          placement="top"
        >
          {{ entry.message }}
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else-if="!loading" :description="localeStore.t('pages.executions.logsEmpty')" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useLocaleStore } from '../stores/locale'
import { getExecution, getExecutionLogs, type Execution, type ExecutionLog } from '../api/executions'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppErrorState from '../components/AppErrorState.vue'

const route = useRoute()
const localeStore = useLocaleStore()
const execution = ref<Execution | null>(null)
const logs = ref<ExecutionLog[]>([])
const loading = ref(false)
const error = ref('')

function statusTagType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    succeeded: 'success',
    running: 'warning',
    failed: 'danger',
    pending: 'info',
  }
  return map[status] ?? 'info'
}

async function loadExecution(executionId: string) {
  loading.value = true
  error.value = ''

  try {
    const [executionDetail, executionLogs] = await Promise.all([
      getExecution(executionId),
      getExecutionLogs(executionId),
    ])
    execution.value = executionDetail
    logs.value = executionLogs
  } catch (err) {
    error.value = err instanceof Error ? err.message : localeStore.t('pages.executions.errors.loadFailed')
  } finally {
    loading.value = false
  }
}

watch(
  () => route.params.id,
  (value) => {
    if (typeof value === 'string' && value) {
      void loadExecution(value)
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.page {
  display: grid;
  gap: 1rem;
}
.error-text {
  color: var(--el-color-danger);
}
</style>
