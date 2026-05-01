<template>
  <div v-loading="loading">
    <el-page-header @back="$router.back()">
      <template #content>Execution Detail</template>
    </el-page-header>

    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon style="margin: 16px 0" />

    <template v-if="execution">
      <el-descriptions :border="true" :column="2" style="margin-top: 16px">
        <el-descriptions-item label="ID">{{ execution.id }}</el-descriptions-item>
        <el-descriptions-item label="Status">
          <el-tag :type="statusTagType(execution.status)" size="small">{{ execution.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Node">{{ execution.nodeId || 'unassigned' }}</el-descriptions-item>
        <el-descriptions-item label="Trigger">{{ execution.triggerSource }}</el-descriptions-item>
        <el-descriptions-item label="Image">{{ execution.image }}</el-descriptions-item>
        <el-descriptions-item label="Command">{{ execution.command.join(' ') || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="execution.startedAt" label="Started">{{ execution.startedAt }}</el-descriptions-item>
        <el-descriptions-item v-if="execution.finishedAt" label="Finished">{{ execution.finishedAt }}</el-descriptions-item>
      </el-descriptions>

      <el-alert
        v-if="execution.errorMessage"
        :title="execution.errorMessage"
        type="error"
        :closable="false"
        show-icon
        style="margin-top: 16px"
      />

      <el-card style="margin-top: 16px">
        <template #header>Logs</template>
        <el-timeline v-if="logs.length">
          <el-timeline-item v-for="entry in logs" :key="entry.id" :timestamp="entry.createdAt" placement="top">
            {{ entry.message }}
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="No logs yet" />
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getExecution, getExecutionLogs, type Execution, type ExecutionLog } from '../api/executions'

const route = useRoute()
const execution = ref<Execution | null>(null)
const logs = ref<ExecutionLog[]>([])
const loading = ref(false)
const error = ref('')

function statusTagType(status: string) {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    pending: 'info',
    running: 'warning',
    succeeded: 'success',
    failed: 'danger',
  }
  return map[status] || 'info'
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
    error.value = err instanceof Error ? err.message : 'failed to load execution'
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
