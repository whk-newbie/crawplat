<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Monitor</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Live overview pulled from the gateway monitor endpoint.
        </p>
      </div>
      <el-button :loading="loading" @click="loadOverview">Refresh</el-button>
    </div>

    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon style="margin-bottom: 16px" />

    <div v-loading="loading">
      <h3 style="margin: 0 0 12px">Executions</h3>
      <el-row :gutter="16" style="margin-bottom: 24px">
        <el-col :span="4" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Total" :value="executions.total" />
          </el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Pending" :value="executions.pending" />
          </el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Running" :value="executions.running" />
          </el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Succeeded" :value="executions.succeeded" />
          </el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Failed" :value="executions.failed" />
          </el-card>
        </el-col>
      </el-row>

      <h3 style="margin: 0 0 12px">Nodes</h3>
      <el-row :gutter="16" style="margin-bottom: 24px">
        <el-col :span="8" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Total Nodes" :value="nodes.total" />
          </el-card>
        </el-col>
        <el-col :span="8" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Online" :value="nodes.online" />
          </el-card>
        </el-col>
        <el-col :span="8" :xs="12">
          <el-card shadow="hover">
            <el-statistic title="Offline" :value="nodes.offline" />
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-collapse>
      <el-collapse-item title="Raw JSON Payload">
        <pre style="margin: 0; white-space: pre-wrap; font-size: 13px">{{ payloadText }}</pre>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { getMonitorOverview } from '../api/monitor'

interface OverviewData {
  executions: { total: number; pending: number; running: number; succeeded: number; failed: number }
  nodes: { total: number; online: number; offline: number }
}

const rawOverview = ref<OverviewData | null>(null)
const loading = ref(true)
const error = ref('')

const executions = reactive({ total: 0, pending: 0, running: 0, succeeded: 0, failed: 0 })
const nodes = reactive({ total: 0, online: 0, offline: 0 })

const payloadText = computed(() => rawOverview.value ? JSON.stringify(rawOverview.value, null, 2) : '')

async function loadOverview() {
  loading.value = true
  error.value = ''

  try {
    const data = await getMonitorOverview() as unknown as OverviewData
    rawOverview.value = data
    if (data.executions) {
      Object.assign(executions, data.executions)
    }
    if (data.nodes) {
      Object.assign(nodes, data.nodes)
    }
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
