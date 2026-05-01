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
          <ul v-if="selectedDetail.recentExecutions.length" class="simple-list">
            <li v-for="execution in selectedDetail.recentExecutions" :key="execution.id">
              <strong>{{ execution.id }}</strong>
              <span>{{ execution.status }}</span>
              <small>{{ execution.spiderId || '-' }}</small>
            </li>
          </ul>
          <p v-else>No recent executions.</p>
        </template>
      </article>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getNodeDetail, listNodes, type NodeDetail, type NodeSummary } from '../api/nodes'

const nodes = ref<NodeSummary[]>([])
const selectedNodeId = ref('')
const selectedDetail = ref<NodeDetail | null>(null)
const loadingList = ref(false)
const loadingDetail = ref(false)
const listError = ref('')
const detailError = ref('')

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
  selectedDetail.value = null
  detailError.value = ''
  loadingDetail.value = true

  try {
    selectedDetail.value = await getNodeDetail(nodeId)
  } catch (err) {
    detailError.value = err instanceof Error ? err.message : 'failed to load node detail'
  } finally {
    loadingDetail.value = false
  }
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

.simple-list {
  display: grid;
  gap: 0.5rem;
  padding-left: 1rem;
}

.simple-list li {
  display: flex;
  gap: 0.5rem;
  align-items: baseline;
}

.error {
  color: #b42318;
}

@media (max-width: 960px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
</style>
