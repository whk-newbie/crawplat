<template>
  <main class="page">
    <section class="card">
      <h1>Execution Detail</h1>
      <p v-if="loading">Loading execution...</p>
      <p v-else-if="error" class="error">{{ error }}</p>
      <template v-else-if="execution">
        <dl class="details">
          <div>
            <dt>ID</dt>
            <dd>{{ execution.id }}</dd>
          </div>
          <div>
            <dt>Status</dt>
            <dd>{{ execution.status }}</dd>
          </div>
          <div>
            <dt>Node</dt>
            <dd>{{ execution.nodeId || 'unassigned' }}</dd>
          </div>
          <div>
            <dt>Trigger</dt>
            <dd>{{ execution.triggerSource }}</dd>
          </div>
          <div>
            <dt>Image</dt>
            <dd>{{ execution.image }}</dd>
          </div>
          <div>
            <dt>Command</dt>
            <dd>{{ execution.command.join(' ') || '-' }}</dd>
          </div>
          <div v-if="execution.errorMessage">
            <dt>Error</dt>
            <dd>{{ execution.errorMessage }}</dd>
          </div>
        </dl>
      </template>
    </section>

    <section class="card">
      <h2>Logs</h2>
      <ul v-if="logs.length" class="logs">
        <li v-for="entry in logs" :key="entry.id">
          <strong>{{ entry.createdAt }}</strong>
          <span>{{ entry.message }}</span>
        </li>
      </ul>
      <p v-else-if="!loading">No logs yet.</p>
    </section>
  </main>
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

<style scoped>
.page {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.card {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 1rem;
}

.details {
  display: grid;
  gap: 0.75rem;
}

.details div {
  display: grid;
  gap: 0.25rem;
}

.logs {
  display: grid;
  gap: 0.75rem;
  padding-left: 1rem;
}

.error {
  color: #b42318;
}
</style>
