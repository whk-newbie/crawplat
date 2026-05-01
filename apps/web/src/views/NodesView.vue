<template>
  <main class="page">
    <section class="card hero">
      <div>
        <p class="eyebrow">Node inventory</p>
        <h1>Nodes</h1>
        <p>Browse node status and inspect heartbeat/execution detail.</p>
      </div>
      <button :disabled="loadingList" @click="loadNodes">
        {{ loadingList ? 'Refreshing...' : 'Refresh' }}
      </button>
    </section>

    <section class="layout">
      <article class="card">
        <h2>Node List</h2>
        <p v-if="loadingList">Loading nodes...</p>
        <p v-else-if="listError" class="error">{{ listError }}</p>
        <p v-else-if="nodes.length === 0">No nodes found.</p>
        <ul v-else class="node-list">
          <li v-for="node in nodes" :key="node.id">
            <button
              :class="{ active: node.id === selectedNodeId }"
              @click="selectNode(node.id)"
            >
              <span>{{ node.name }}</span>
              <small>{{ node.status }}</small>
            </button>
          </li>
        </ul>
      </article>

      <article class="card">
        <h2>Node Detail</h2>
        <p v-if="loadingDetail">Loading detail...</p>
        <p v-else-if="detailError" class="error">{{ detailError }}</p>
        <p v-else-if="!selectedDetail">Select a node to view detail.</p>
        <template v-else>
          <dl class="details">
            <div>
              <dt>ID</dt>
              <dd>{{ selectedDetail.id }}</dd>
            </div>
            <div>
              <dt>Status</dt>
              <dd>{{ selectedDetail.status }}</dd>
            </div>
            <div>
              <dt>Last Seen</dt>
              <dd>{{ selectedDetail.lastSeenAt || '-' }}</dd>
            </div>
          </dl>

          <h3>Capabilities</h3>
          <div class="chips">
            <span v-for="capability in selectedDetail.capabilities" :key="capability">
              {{ capability }}
            </span>
            <p v-if="selectedDetail.capabilities.length === 0">No capabilities.</p>
          </div>

          <h3>Heartbeats</h3>
          <ul v-if="selectedDetail.heartbeats.length" class="simple-list">
            <li v-for="(heartbeat, index) in selectedDetail.heartbeats" :key="`${heartbeat.seenAt}-${index}`">
              {{ heartbeat.seenAt }}
              <small v-if="heartbeat.status">({{ heartbeat.status }})</small>
            </li>
          </ul>
          <p v-else>No heartbeat history yet.</p>

          <h3>Recent Executions</h3>
          <div class="toolbar wrap">
            <label>
              Status
              <select v-model="executionStatus" :disabled="loadingDetail" @change="reloadSelectedDetail">
                <option value="">all</option>
                <option value="queued">queued</option>
                <option value="running">running</option>
                <option value="succeeded">succeeded</option>
                <option value="failed">failed</option>
                <option value="cancelled">cancelled</option>
              </select>
            </label>
            <label>
              From
              <input
                v-model="executionFromInput"
                :disabled="loadingDetail"
                type="datetime-local"
                @change="resetExecutionPageAndReload"
              />
            </label>
            <label>
              To
              <input
                v-model="executionToInput"
                :disabled="loadingDetail"
                type="datetime-local"
                @change="resetExecutionPageAndReload"
              />
            </label>
            <label>
              Limit
              <select v-model.number="executionLimit" :disabled="loadingDetail" @change="resetExecutionPageAndReload">
                <option :value="10">10</option>
                <option :value="20">20</option>
                <option :value="50">50</option>
              </select>
            </label>
            <button :disabled="loadingDetail || executionOffset === 0" @click="prevExecutionsPage">Prev</button>
            <button :disabled="loadingDetail" @click="nextExecutionsPage">Next</button>
            <small>offset: {{ executionOffset }}</small>
          </div>
          <ul v-if="selectedDetail.recentExecutions.length" class="simple-list">
            <li v-for="execution in selectedDetail.recentExecutions" :key="execution.id">
              <strong>{{ execution.id }}</strong>
              <span>{{ execution.status }}</span>
              <small>{{ execution.spiderId || '-' }}</small>
            </li>
          </ul>
          <p v-else>No recent executions.</p>

          <h3>Online Sessions</h3>
          <div class="toolbar wrap">
            <label>
              Limit
              <input
                v-model.number="sessionLimit"
                :disabled="loadingSessions"
                min="1"
                step="1"
                type="number"
              />
            </label>
            <label>
              Gap Seconds
              <input
                v-model.number="sessionGapSeconds"
                :disabled="loadingSessions"
                min="1"
                step="1"
                type="number"
              />
            </label>
            <button :disabled="loadingSessions" @click="reloadSelectedSessions">
              {{ loadingSessions ? 'Loading...' : 'Load Sessions' }}
            </button>
          </div>
          <div class="session-summary">
            <article class="session-stat">
              <small>sessions</small>
              <strong>{{ sessionSummary.totalSessions }}</strong>
            </article>
            <article class="session-stat">
              <small>heartbeats</small>
              <strong>{{ sessionSummary.totalHeartbeatCount }}</strong>
            </article>
            <article class="session-stat">
              <small>online seconds</small>
              <strong>{{ sessionSummary.totalOnlineDurationSeconds }}</strong>
            </article>
          </div>
          <p v-if="sessionsError" class="error">{{ sessionsError }}</p>
          <ul v-else-if="sessions.length" class="simple-list">
            <li v-for="(session, index) in sessions" :key="`${session.startedAt}-${index}`">
              <strong>{{ session.startedAt }}</strong>
              <span>~ {{ session.endedAt || 'now' }}</span>
              <small>heartbeats: {{ session.heartbeatCount ?? '-' }}</small>
              <small>duration: {{ session.durationSeconds ?? '-' }}s</small>
            </li>
          </ul>
          <p v-else>No session history yet.</p>
        </template>
      </article>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
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

function normalizeDateInput(value: string) {
  if (!value.trim()) {
    return undefined
  }
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return undefined
  }
  return parsed.toISOString()
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
  } finally {
    loadingList.value = false
  }
}

async function selectNode(nodeId: string) {
  selectedNodeId.value = nodeId
  executionOffset.value = 0
  selectedDetail.value = null
  sessions.value = []
  sessionSummary.value = {
    totalSessions: 0,
    totalHeartbeatCount: 0,
    totalOnlineDurationSeconds: 0,
  }
  await Promise.all([reloadSelectedDetail(), reloadSelectedSessions()])
}

async function reloadSelectedDetail() {
  if (!selectedNodeId.value) {
    return
  }
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
  if (!selectedNodeId.value) {
    return
  }
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
  } finally {
    loadingSessions.value = false
  }
}

async function resetExecutionPageAndReload() {
  executionOffset.value = 0
  await reloadSelectedDetail()
}

async function prevExecutionsPage() {
  if (executionOffset.value === 0) {
    return
  }
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

.layout {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(16rem, 22rem) minmax(0, 1fr);
}

.node-list {
  display: grid;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.node-list button {
  width: 100%;
  border: 1px solid #d0d7de;
  border-radius: 8px;
  background: #f6f8fa;
  padding: 0.5rem 0.75rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  text-align: left;
  cursor: pointer;
}

.node-list button.active {
  border-color: #0969da;
  background: #ddf4ff;
}

.details {
  display: grid;
  gap: 0.5rem;
}

.details div {
  display: grid;
  gap: 0.25rem;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.chips span {
  padding: 0.2rem 0.5rem;
  border: 1px solid #d0d7de;
  border-radius: 999px;
  background: #f6f8fa;
  font-size: 0.85rem;
}

.session-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
  margin: 0.5rem 0;
}

.session-stat {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 0.5rem 0.6rem;
  background: #f6f8fa;
}

.session-stat small {
  display: block;
  color: #57606a;
  text-transform: uppercase;
  font-size: 0.7rem;
}

.simple-list {
  display: grid;
  gap: 0.5rem;
  padding-left: 1rem;
}

.simple-list li {
  display: flex;
  gap: 0.5rem;
  align-items: baseline;
  flex-wrap: wrap;
}

.error {
  color: #b42318;
}

.toolbar {
  display: flex;
  gap: 0.75rem;
  align-items: end;
  margin-bottom: 0.75rem;
}

.toolbar label {
  display: grid;
  gap: 0.25rem;
  font-size: 0.85rem;
}

.toolbar select,
.toolbar input {
  min-width: 6rem;
}

.toolbar.wrap {
  flex-wrap: wrap;
}

@media (max-width: 960px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
</style>
