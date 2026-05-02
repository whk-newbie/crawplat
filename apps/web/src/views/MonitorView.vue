<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Monitor</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Live overview, alert rules, and recent alert delivery events.
        </p>
      </div>
      <el-button :loading="loading" @click="loadAll">Refresh</el-button>
    </div>

    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon style="margin-bottom: 16px" />

    <div v-loading="loading">
      <h3 style="margin: 0 0 12px">Executions</h3>
      <el-row :gutter="16" style="margin-bottom: 24px">
        <el-col :span="4" :xs="12">
          <el-card shadow="hover"><el-statistic title="Total" :value="executions.total" /></el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover"><el-statistic title="Pending" :value="executions.pending" /></el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover"><el-statistic title="Running" :value="executions.running" /></el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover"><el-statistic title="Succeeded" :value="executions.succeeded" /></el-card>
        </el-col>
        <el-col :span="4" :xs="12">
          <el-card shadow="hover"><el-statistic title="Failed" :value="executions.failed" /></el-card>
        </el-col>
      </el-row>

      <h3 style="margin: 0 0 12px">Nodes</h3>
      <el-row :gutter="16" style="margin-bottom: 24px">
        <el-col :span="8" :xs="12"><el-card shadow="hover"><el-statistic title="Total Nodes" :value="nodes.total" /></el-card></el-col>
        <el-col :span="8" :xs="12"><el-card shadow="hover"><el-statistic title="Online" :value="nodes.online" /></el-card></el-col>
        <el-col :span="8" :xs="12"><el-card shadow="hover"><el-statistic title="Offline" :value="nodes.offline" /></el-card></el-col>
      </el-row>

      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center">
            <span>Alert Rules</span>
          </div>
        </template>

        <el-form :inline="true" :model="createForm" @submit.prevent="submitCreateRule">
          <el-form-item label="Name"><el-input v-model="createForm.name" placeholder="Rule name" /></el-form-item>
          <el-form-item label="Type">
            <el-select v-model="createForm.ruleType" style="width: 180px">
              <el-option label="Execution Failed" value="execution_failed" />
              <el-option label="Node Offline" value="node_offline" />
            </el-select>
          </el-form-item>
          <el-form-item label="Webhook"><el-input v-model="createForm.webhookUrl" placeholder="https://..." style="width: 320px" /></el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="creatingRule" @click="submitCreateRule">Create Rule</el-button>
          </el-form-item>
        </el-form>

        <el-table :data="rules" size="small" style="margin-top: 8px">
          <el-table-column prop="name" label="Name" min-width="160" />
          <el-table-column prop="ruleType" label="Type" width="160" />
          <el-table-column prop="webhookUrl" label="Webhook" min-width="260" show-overflow-tooltip />
          <el-table-column label="Enabled" width="120">
            <template #default="{ row }">
              <el-switch :model-value="row.enabled" :loading="updatingRuleId === row.id" @change="toggleRule(row.id, $event)" />
            </template>
          </el-table-column>
          <el-table-column prop="timeoutSeconds" label="Timeout(s)" width="120" />
          <el-table-column prop="cooldownSeconds" label="Cooldown(s)" width="130" />
        </el-table>
      </el-card>

      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header><span>Recent Alert Events</span></template>
        <el-table :data="events" size="small">
          <el-table-column prop="createdAt" label="Time" min-width="180" />
          <el-table-column prop="ruleType" label="Rule Type" width="160" />
          <el-table-column prop="entityId" label="Entity" min-width="140" />
          <el-table-column prop="deliveryStatus" label="Status" width="110" />
          <el-table-column prop="webhookStatusCode" label="HTTP" width="90" />
          <el-table-column prop="errorMessage" label="Error" min-width="240" show-overflow-tooltip />
        </el-table>
      </el-card>
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
import { ElMessage } from 'element-plus'
import {
  createAlertRule,
  getMonitorOverview,
  listAlertEvents,
  listAlertRules,
  updateAlertRule,
  type AlertEvent,
  type AlertRule,
} from '../api/monitor'

interface OverviewData {
  executions: { total: number; pending: number; running: number; succeeded: number; failed: number }
  nodes: { total: number; online: number; offline: number }
}

const rawOverview = ref<OverviewData | null>(null)
const loading = ref(true)
const error = ref('')
const creatingRule = ref(false)
const updatingRuleId = ref('')

const rules = ref<AlertRule[]>([])
const events = ref<AlertEvent[]>([])

const createForm = reactive({
  name: '',
  ruleType: 'execution_failed' as 'execution_failed' | 'node_offline',
  webhookUrl: '',
})

const executions = reactive({ total: 0, pending: 0, running: 0, succeeded: 0, failed: 0 })
const nodes = reactive({ total: 0, online: 0, offline: 0 })

const payloadText = computed(() => rawOverview.value ? JSON.stringify(rawOverview.value, null, 2) : '')

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    const [overview, ruleItems, eventPayload] = await Promise.all([
      getMonitorOverview() as Promise<OverviewData>,
      listAlertRules(),
      listAlertEvents(20, 0),
    ])

    rawOverview.value = overview
    if (overview.executions) Object.assign(executions, overview.executions)
    if (overview.nodes) Object.assign(nodes, overview.nodes)

    rules.value = ruleItems
    events.value = eventPayload.items
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to load monitor overview'
  } finally {
    loading.value = false
  }
}

async function submitCreateRule() {
  if (!createForm.name.trim() || !createForm.webhookUrl.trim()) {
    ElMessage.error('name and webhook are required')
    return
  }
  creatingRule.value = true
  try {
    await createAlertRule({
      name: createForm.name.trim(),
      ruleType: createForm.ruleType,
      webhookUrl: createForm.webhookUrl.trim(),
    })
    createForm.name = ''
    createForm.webhookUrl = ''
    await loadAll()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create alert rule')
  } finally {
    creatingRule.value = false
  }
}

async function toggleRule(id: string, enabled: boolean | string | number) {
  updatingRuleId.value = id
  try {
    await updateAlertRule(id, { enabled: Boolean(enabled) })
    await loadAll()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to update alert rule')
  } finally {
    updatingRuleId.value = ''
  }
}

onMounted(() => {
  void loadAll()
})
</script>
