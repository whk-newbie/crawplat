<template>
  <main class="page">
    <section class="hero card">
      <div>
        <p class="eyebrow">Control plane</p>
        <h1>Monitor</h1>
        <p>Live overview pulled from the gateway monitor endpoint.</p>
      </div>
      <button class="refresh" :disabled="loading" @click="loadOverview">
        {{ loading ? 'Refreshing...' : 'Refresh' }}
      </button>
    </section>

    <section class="card">
      <p v-if="loading">Loading overview...</p>
      <p v-else-if="error" class="error">{{ error }}</p>
      <template v-else>
        <dl v-if="counterEntries.length" class="counters">
          <div v-for="item in counterEntries" :key="item.label">
            <dt>{{ item.label }}</dt>
            <dd>{{ item.value }}</dd>
          </div>
        </dl>
        <p v-else>No counters returned yet.</p>
      </template>
    </section>

    <section class="card">
      <h2>Overview payload</h2>
      <pre v-if="overview" class="payload">{{ payloadText }}</pre>
      <p v-else>No overview loaded.</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getMonitorOverview, type MonitorOverview } from '../api/monitor'

const overview = ref<MonitorOverview | null>(null)
const loading = ref(true)
const error = ref('')

const counterDefinitions = [
  ['activeExecutions', 'Active executions'],
  ['runningExecutions', 'Running executions'],
  ['queuedExecutions', 'Queued executions'],
  ['failedExecutions', 'Failed executions'],
  ['healthyNodes', 'Healthy nodes'],
  ['activeNodes', 'Active nodes'],
  ['nodesOnline', 'Nodes online'],
  ['datasourceChecks', 'Datasource checks'],
]

const counterEntries = computed(() =>
  counterDefinitions
    .map(([key, label]) => {
      const value = overview.value?.[key]
      return typeof value === 'number' || typeof value === 'string'
        ? { label, value }
        : null
    })
    .filter((item): item is { label: string; value: string | number } => item !== null),
)

const payloadText = computed(() => JSON.stringify(overview.value, null, 2))

async function loadOverview() {
  loading.value = true
  error.value = ''

  try {
    overview.value = await getMonitorOverview()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to load monitor overview'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadOverview()
})
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

.hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.eyebrow {
  margin: 0 0 0.25rem;
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #57606a;
}

.refresh {
  white-space: nowrap;
}

.counters {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
}

.counters div {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 0.75rem;
  background: #f6f8fa;
}

.counters dt {
  font-size: 0.8rem;
  color: #57606a;
}

.counters dd {
  margin: 0.25rem 0 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.payload {
  overflow: auto;
  margin: 0;
  padding: 0.75rem;
  border-radius: 8px;
  background: #f6f8fa;
}

.error {
  color: #b42318;
}
</style>
