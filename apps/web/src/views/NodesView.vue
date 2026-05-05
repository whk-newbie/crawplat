<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { NodeSummary, NodeDetail, NodeRecentExecution, NodeSession } from '../api/nodes'
import { getNodeDetail, getNodeSessions, listNodes, type NodeSessionsResult } from '../api/nodes'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()

const nodes = ref<NodeSummary[]>([])
const loading = ref(false)
const loadError = ref('')

const selectedNode = ref<NodeDetail | null>(null)
const sessions = ref<NodeSessionsResult | null>(null)
const detailLoading = ref(false)
const detailError = ref('')
const showDetail = ref(false)

const executionStatusFilter = ref('')

async function loadNodes() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listNodes()
    nodes.value = res.items ?? []
  } catch {
    loadError.value = localeStore.t('pages.nodes.errors.loadFailed')
    nodes.value = []
  } finally {
    loading.value = false
  }
}

async function openDetail(node: NodeSummary) {
  selectedNode.value = null
  sessions.value = null
  detailLoading.value = true
  detailError.value = ''
  showDetail.value = true

  try {
    const query = executionStatusFilter.value
      ? { executionStatus: executionStatusFilter.value }
      : {}
    const [detail, nodeSessions] = await Promise.all([
      getNodeDetail(node.id, query),
      getNodeSessions(node.id),
    ])
    selectedNode.value = detail
    sessions.value = nodeSessions
  } catch {
    detailError.value = localeStore.t('pages.nodes.errors.loadDetailFailed')
  } finally {
    detailLoading.value = false
  }
}

function statusTagType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    online: 'success',
    warning: 'warning',
    offline: 'danger',
  }
  return map[status] ?? 'info'
}

function formatDateTime(value: string | undefined): string {
  if (!value) return '-'
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}

onMounted(loadNodes)
</script>

<template>
  <main class="nodes-page">
    <div class="page-header">
      <h1>{{ localeStore.t('pages.nodes.title') }}</h1>
      <el-button @click="loadNodes">{{ localeStore.t('common.actions.refresh') }}</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="nodes"
      stripe
      :empty-text="localeStore.t('pages.nodes.empty')"
      @row-click="(row: NodeSummary) => openDetail(row)"
    >
      <el-table-column prop="name" :label="localeStore.t('pages.nodes.name')" min-width="140" />
      <el-table-column :label="localeStore.t('pages.nodes.status')" width="100">
        <template #default="scope">
          <el-tag :type="statusTagType(scope.row.status)" size="small">
            {{ scope.row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="localeStore.t('pages.nodes.capabilities')" min-width="160">
        <template #default="scope">
          <template v-if="scope.row.capabilities?.length">
            <el-tag
              v-for="cap in scope.row.capabilities"
              :key="cap"
              size="small"
              type="info"
              style="margin-right: 4px"
            >
              {{ cap }}
            </el-tag>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column :label="localeStore.t('pages.nodes.lastSeen')" min-width="180">
        <template #default="scope">
          {{ formatDateTime(scope.row.lastSeenAt) }}
        </template>
      </el-table-column>
    </el-table>

    <el-alert
      v-if="loadError"
      :title="loadError"
      type="error"
      show-icon
      class="error-alert"
    />

    <!-- Detail Drawer -->
    <el-drawer
      v-model="showDetail"
      :title="selectedNode ? `${localeStore.t('pages.nodes.detail.title')} - ${selectedNode.name}` : localeStore.t('pages.nodes.detail.title')"
      size="600px"
    >
      <div v-loading="detailLoading">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          class="error-alert"
        />

        <template v-if="selectedNode">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item :label="localeStore.t('pages.nodes.name')">
              {{ selectedNode.name }}
            </el-descriptions-item>
            <el-descriptions-item :label="localeStore.t('pages.nodes.status')">
              <el-tag :type="statusTagType(selectedNode.status)" size="small">
                {{ selectedNode.status }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="localeStore.t('pages.nodes.lastSeen')">
              {{ formatDateTime(selectedNode.lastSeenAt) }}
            </el-descriptions-item>
          </el-descriptions>

          <h3 class="section-title">{{ localeStore.t('pages.nodes.detail.heartbeats') }}</h3>
          <el-timeline v-if="selectedNode.heartbeats?.length">
            <el-timeline-item
              v-for="hb in selectedNode.heartbeats"
              :key="hb.seenAt"
              :timestamp="formatDateTime(hb.seenAt)"
              placement="top"
            >
              Status: {{ hb.status ?? 'unknown' }}
            </el-timeline-item>
          </el-timeline>
          <el-empty v-else :description="localeStore.t('common.status.empty')" />

          <h3 class="section-title">
            {{ localeStore.t('pages.nodes.detail.recentExecutions') }}
            <el-select
              v-model="executionStatusFilter"
              size="small"
              style="width: 130px; margin-left: 0.5rem"
              :placeholder="localeStore.t('pages.nodes.detail.statusFilter')"
              @change="(val: string) => openDetail({ id: selectedNode!.id, name: selectedNode!.name, status: selectedNode!.status, capabilities: selectedNode!.capabilities ?? [], lastSeenAt: selectedNode!.lastSeenAt })"
            >
              <el-option label="All" value="" />
              <el-option label="Pending" value="pending" />
              <el-option label="Running" value="running" />
              <el-option label="Succeeded" value="succeeded" />
              <el-option label="Failed" value="failed" />
            </el-select>
          </h3>
          <el-table
            v-if="selectedNode.recentExecutions?.length"
            :data="selectedNode.recentExecutions"
            size="small"
          >
            <el-table-column prop="id" label="ID" width="100" />
            <el-table-column :label="localeStore.t('pages.nodes.status')" width="100">
              <template #default="scope">
                <el-tag :type="statusTagType(scope.row.status)" size="small">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="localeStore.t('pages.nodes.detail.startedAt')" min-width="160">
              <template #default="scope">
                {{ formatDateTime(scope.row.startedAt) }}
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else :description="localeStore.t('common.status.empty')" />

          <h3 class="section-title">{{ localeStore.t('pages.nodes.detail.sessions') }}</h3>
          <template v-if="sessions">
            <el-descriptions :column="2" size="small" border style="margin-bottom: 1rem">
              <el-descriptions-item :label="localeStore.t('pages.nodes.detail.sessionCount')">
                {{ sessions.summary.totalSessions }}
              </el-descriptions-item>
              <el-descriptions-item :label="localeStore.t('pages.nodes.detail.totalDuration')">
                {{ sessions.summary.totalOnlineDurationSeconds }}
              </el-descriptions-item>
            </el-descriptions>
            <el-table v-if="sessions.sessions.length" :data="sessions.sessions" size="small">
              <el-table-column :label="localeStore.t('pages.nodes.detail.startedAt')" min-width="160">
                <template #default="scope">
                  {{ formatDateTime(scope.row.startedAt) }}
                </template>
              </el-table-column>
              <el-table-column :label="localeStore.t('pages.nodes.detail.endedAt')" min-width="160">
                <template #default="scope">
                  {{ formatDateTime(scope.row.endedAt) }}
                </template>
              </el-table-column>
              <el-table-column :label="localeStore.t('pages.nodes.detail.durationSeconds')" width="120">
                <template #default="scope">
                  {{ scope.row.durationSeconds ?? '-' }}
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else :description="localeStore.t('common.status.empty')" />
          </template>
        </template>
      </div>
    </el-drawer>
  </main>
</template>

<style scoped>
.nodes-page {
  padding: 0;
}

.page-header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.page-header h1 {
  margin: 0;
}

.error-alert {
  margin-bottom: 1rem;
  margin-top: 1rem;
}

.section-title {
  align-items: center;
  display: flex;
  margin: 1.5rem 0 0.75rem;
}
</style>
