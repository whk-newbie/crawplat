<template>
  <div class="page">
    <el-card>
      <template #header>
        <div class="hero">
          <div>
            <h1>{{ localeStore.t('pages.monitor.title') }}</h1>
            <p class="subtitle">Live overview pulled from the gateway monitor endpoint.</p>
          </div>
          <el-button :loading="loading" @click="loadOverview">
            {{ loading ? localeStore.t('common.state.loading') : localeStore.t('common.actions.refresh') }}
          </el-button>
        </div>
      </template>
    </el-card>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="loadOverview" />

    <el-card v-if="!loading && !error">
      <el-row v-if="counterEntries.length" :gutter="16">
        <el-col v-for="item in counterEntries" :key="item.label" :span="6" :xs="12">
          <el-statistic :title="item.label" :value="item.value" />
        </el-col>
      </el-row>
      <el-empty v-else description="No counters returned yet." />
    </el-card>

    <el-card v-if="!loading && !error">
      <template #header>
        <h2>Raw overview payload</h2>
      </template>
      <pre v-if="overview" class="payload">{{ payloadText }}</pre>
      <el-empty v-else description="No overview loaded." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useLocaleStore } from '../stores/locale'
import { getMonitorOverview, type MonitorOverview } from '../api/monitor'
import AppLoadingState from '../components/AppLoadingState.vue'
import AppErrorState from '../components/AppErrorState.vue'

const localeStore = useLocaleStore()
const overview = ref<MonitorOverview | null>(null)
const loading = ref(true)
const error = ref('')

const counterDefinitions = [
  ['executions.total', 'Total executions'],
  ['executions.pending', 'Pending executions'],
  ['executions.running', 'Running executions'],
  ['executions.failed', 'Failed executions'],
  ['executions.succeeded', 'Succeeded executions'],
  ['nodes.total', 'Total nodes'],
  ['nodes.online', 'Nodes online'],
  ['nodes.offline', 'Nodes offline'],
]

const counterEntries = computed(() =>
  counterDefinitions
    .map(([path, label]) => {
      const value = path.split('.').reduce<unknown>((current, segment) => {
        if (current && typeof current === 'object' && segment in current) {
          return (current as Record<string, unknown>)[segment]
        }
        return undefined
      }, overview.value)
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
}
.hero {
  align-items: flex-start;
  display: flex;
  justify-content: space-between;
  width: 100%;
}
.subtitle {
  color: var(--el-text-color-secondary);
  margin: 4px 0 0;
}
.payload {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  margin: 0;
  overflow: auto;
  padding: 0.75rem;
}
</style>
