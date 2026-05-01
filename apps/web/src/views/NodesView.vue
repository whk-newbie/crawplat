<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Nodes</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Browse node status and inspect heartbeat/execution detail.
        </p>
      </div>
      <el-button :loading="loadingList" @click="loadNodes">Refresh</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="8" :xs="24">
        <el-card>
          <template #header>Node List</template>
          <el-table
            v-loading="loadingList"
            :data="nodes"
            highlight-current-row
            @current-change="onNodeSelect"
            style="width: 100%"
          >
            <el-table-column prop="name" label="Name" />
            <el-table-column prop="status" label="Status" width="90">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty description="No nodes found" />
            </template>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="16" :xs="24">
        <el-card v-loading="loadingDetail">
          <template #header>Node Detail</template>

          <el-empty v-if="!selectedDetail && !detailError" description="Select a node to view details" />
          <el-alert v-if="detailError" :title="detailError" type="error" :closable="false" show-icon />

          <template v-if="selectedDetail">
            <el-descriptions :border="true" :column="2" style="margin-bottom: 16px">
              <el-descriptions-item label="ID">{{ selectedDetail.id }}</el-descriptions-item>
              <el-descriptions-item label="Status">
                <el-tag :type="selectedDetail.status === 'online' ? 'success' : 'danger'" size="small">{{ selectedDetail.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="Last Seen">{{ selectedDetail.lastSeenAt || '-' }}</el-descriptions-item>
            </el-descriptions>

            <h4 style="margin: 16px 0 8px">Capabilities</h4>
            <div v-if="selectedDetail.capabilities.length" style="display: flex; flex-wrap: wrap; gap: 6px">
              <el-tag v-for="cap in selectedDetail.capabilities" :key="cap" size="small">{{ cap }}</el-tag>
            </div>
            <p v-else style="color: var(--el-text-color-secondary)">No capabilities.</p>

            <h4 style="margin: 16px 0 8px">Heartbeats</h4>
            <el-timeline v-if="selectedDetail.heartbeats.length">
              <el-timeline-item
                v-for="(hb, idx) in selectedDetail.heartbeats"
                :key="`${hb.seenAt}-${idx}`"
                :timestamp="hb.seenAt"
                placement="top"
              >
                {{ hb.status || 'heartbeat' }}
              </el-timeline-item>
            </el-timeline>
            <el-empty v-else description="No heartbeat history" :image-size="60" />

            <h4 style="margin: 16px 0 8px">Recent Executions</h4>
            <div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: end; margin-bottom: 12px">
              <el-select v-model="executionStatus" placeholder="Status" style="width: 120px" @change="reloadSelectedDetail">
                <el-option label="All" value="" />
                <el-option label="Queued" value="queued" />
                <el-option label="Running" value="running" />
                <el-option label="Succeeded" value="succeeded" />
                <el-option label="Failed" value="failed" />
                <el-option label="Cancelled" value="cancelled" />
              </el-select>
              <el-date-picker v-model="executionFromInput" type="datetime" placeholder="From" style="width: 180px" @change="resetExecutionPageAndReload" />
              <el-date-picker v-model="executionToInput" type="datetime" placeholder="To" style="width: 180px" @change="resetExecutionPageAndReload" />
              <el-select v-model="executionLimit" style="width: 80px" @change="resetExecutionPageAndReload">
                <el-option :value="10" label="10" />
                <el-option :value="20" label="20" />
                <el-option :value="50" label="50" />
              </el-select>
              <el-button-group>
                <el-button :disabled="loadingDetail || executionOffset === 0" @click="prevExecutionsPage">Prev</el-button>
                <el-button :disabled="loadingDetail" @click="nextExecutionsPage">Next</el-button>
              </el-button-group>
              <span style="font-size: 12px; color: var(--el-text-color-secondary)">offset: {{ executionOffset }}</span>
            </div>
            <el-table v-if="selectedDetail.recentExecutions.length" :data="selectedDetail.recentExecutions" size="small">
              <el-table-column prop="id" label="ID" />
              <el-table-column prop="spiderId" label="Spider" />
              <el-table-column prop="status" label="Status" width="100">
                <template #default="{ row }">
                  <el-tag :type="execStatusType(row.status)" size="small">{{ row.status }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="No recent executions" :image-size="60" />

            <h4 style="margin: 16px 0 8px">Online Sessions</h4>
            <div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: end; margin-bottom: 12px">
              <div>
                <span style="font-size: 12px; color: var(--el-text-color-secondary); display: block; margin-bottom: 4px">Limit</span>
                <el-input-number v-model="sessionLimit" :min="1" size="small" style="width: 100px" />
              </div>
              <div>
                <span style="font-size: 12px; color: var(--el-text-color-secondary); display: block; margin-bottom: 4px">Gap (s)</span>
                <el-input-number v-model="sessionGapSeconds" :min="1" size="small" style="width: 100px" />
              </div>
              <el-button :loading="loadingSessions" size="small" @click="reloadSelectedSessions">Load Sessions</el-button>
            </div>
            <el-row :gutter="12" style="margin-bottom: 12px">
              <el-col :span="8">
                <el-card shadow="never"><el-statistic title="Sessions" :value="sessionSummary.totalSessions" /></el-card>
              </el-col>
              <el-col :span="8">
                <el-card shadow="never"><el-statistic title="Heartbeats" :value="sessionSummary.totalHeartbeatCount" /></el-card>
              </el-col>
              <el-col :span="8">
                <el-card shadow="never"><el-statistic title="Online (s)" :value="sessionSummary.totalOnlineDurationSeconds" /></el-card>
              </el-col>
            </el-row>
            <el-alert v-if="sessionsError" :title="sessionsError" type="error" :closable="false" show-icon />
            <el-table v-else-if="sessions.length" :data="sessions" size="small">
              <el-table-column prop="startedAt" label="Started" />
              <el-table-column label="Ended">
                <template #default="{ row }">{{ row.endedAt || 'now' }}</template>
              </el-table-column>
              <el-table-column prop="heartbeatCount" label="Heartbeats" width="100" />
              <el-table-column label="Duration" width="100">
                <template #default="{ row }">{{ row.durationSeconds ?? '-' }}s</template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="No session history" :image-size="60" />
          </template>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getNodeDetail,
  getNodeSessions,
  listNodes,
  type NodeDetail,
  type NodeSession,
  type NodeSessionsSummary,
  type NodeSummary,
} from '../api/nodes'

const nodes = ref<NodeSummary[]>([])
const selectedNodeId = ref('')
const selectedDetail = ref<NodeDetail | null>(null)
const loadingList = ref(false)
const loadingDetail = ref(false)
const loadingSessions = ref(false)
const listError = ref('')
const detailError = ref('')
const sessionsError = ref('')
const sessions = ref<NodeSession[]>([])
const executionLimit = ref(20)
const executionOffset = ref(0)
const executionStatus = ref('')
const executionFromInput = ref('')
const executionToInput = ref('')
const sessionLimit = ref(20)
const sessionGapSeconds = ref(90)
const sessionSummary = ref<NodeSessionsSummary>({
  totalSessions: 0,
  totalHeartbeatCount: 0,
  totalOnlineDurationSeconds: 0,
})

function execStatusType(status: string) {
  const map: Record<string, string> = { pending: 'info', running: 'warning', succeeded: 'success', failed: 'danger' }
  return map[status] || 'info'
}

function normalizeDateInput(value: string) {
  if (!value || !value.trim()) {
    return undefined
  }
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return undefined
  }
  return parsed.toISOString()
}

function onNodeSelect(row: NodeSummary | null) {
  if (row) {
    void selectNode(row.id)
  }
}

async function loadNodes() {
  loadingList.value = true
  listError.value = ''
  try {
    const result = await listNodes()
    nodes.value = result
    if (!selectedNodeId.value && result.length > 0) {
      await selectNode(result[0].id)
    }
  } catch (err) {
    listError.value = err instanceof Error ? err.message : 'failed to load nodes'
    ElMessage.error(listError.value)
  } finally {
    loadingList.value = false
  }
}

async function selectNode(nodeId: string) {
  selectedNodeId.value = nodeId
  executionOffset.value = 0
  selectedDetail.value = null
  sessions.value = []
  sessionSummary.value = { totalSessions: 0, totalHeartbeatCount: 0, totalOnlineDurationSeconds: 0 }
  await Promise.all([reloadSelectedDetail(), reloadSelectedSessions()])
}

async function reloadSelectedDetail() {
  if (!selectedNodeId.value) return
  detailError.value = ''
  loadingDetail.value = true
  try {
    selectedDetail.value = await getNodeDetail(selectedNodeId.value, {
      executionLimit: executionLimit.value,
      executionOffset: executionOffset.value,
      executionStatus: executionStatus.value.trim() || undefined,
      executionFrom: normalizeDateInput(executionFromInput.value),
      executionTo: normalizeDateInput(executionToInput.value),
    })
  } catch (err) {
    detailError.value = err instanceof Error ? err.message : 'failed to load node detail'
  } finally {
    loadingDetail.value = false
  }
}

async function reloadSelectedSessions() {
  if (!selectedNodeId.value) return
  sessionsError.value = ''
  loadingSessions.value = true
  try {
    const result = await getNodeSessions(selectedNodeId.value, {
      limit: Math.max(1, sessionLimit.value || 20),
      gapSeconds: Math.max(1, sessionGapSeconds.value || 90),
    })
    sessions.value = result.sessions
    sessionSummary.value = result.summary
  } catch (err) {
    sessionsError.value = err instanceof Error ? err.message : 'failed to load sessions'
    ElMessage.error(sessionsError.value)
  } finally {
    loadingSessions.value = false
  }
}

async function resetExecutionPageAndReload() {
  executionOffset.value = 0
  await reloadSelectedDetail()
}

async function prevExecutionsPage() {
  if (executionOffset.value === 0) return
  executionOffset.value = Math.max(0, executionOffset.value - executionLimit.value)
  await reloadSelectedDetail()
}

async function nextExecutionsPage() {
  executionOffset.value += executionLimit.value
  await reloadSelectedDetail()
}

onMounted(() => {
  void loadNodes()
})
</script>
