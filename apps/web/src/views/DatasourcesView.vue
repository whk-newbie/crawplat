<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Datasources</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Manage data connections and run connectivity checks or row preview.
        </p>
      </div>
      <div>
        <el-button :loading="loading" @click="loadDatasources">Refresh</el-button>
        <el-button type="primary" @click="openCreateDialog">Create Datasource</el-button>
      </div>
    </div>

    <el-form :inline="true" style="margin-bottom: 12px">
      <el-form-item label="Project ID">
        <el-input v-model.trim="projectFilter" placeholder="project-1" style="width: 200px" />
      </el-form-item>
      <el-form-item>
        <el-button :loading="loading" @click="loadDatasources">Query</el-button>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="datasources" stripe style="margin-bottom: 16px">
      <el-table-column prop="name" label="Name" />
      <el-table-column prop="type" label="Type" width="120">
        <template #default="{ row }">
          <el-tag :type="dsTypeTag(row.type)" size="small">{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="projectId" label="Project" />
      <el-table-column label="Readonly" width="100">
        <template #default="{ row }">
          <el-tag :type="row.readonly ? 'info' : 'warning'" size="small">{{ row.readonly ? 'readonly' : 'writable' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Config">
        <template #default="{ row }">
          <span v-for="entry in configEntries(row.config)" :key="entry.key" style="margin-right: 8px">
            <strong>{{ entry.key }}:</strong> {{ entry.value }}
          </span>
          <span v-if="configEntries(row.config).length === 0" style="color: var(--el-text-color-secondary)">-</span>
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="160">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" :loading="testingId === row.id" @click="runTest(row.id)">Test</el-button>
            <el-button size="small" :loading="previewingId === row.id" @click="runPreview(row.id)">Preview</el-button>
          </el-button-group>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty v-if="!loading" description="No datasources found" />
      </template>
    </el-table>

    <el-card v-if="lastTestResult || lastPreviewResult || actionError">
      <template #header>Action Result</template>
      <el-alert v-if="actionError" :title="actionError" type="error" :closable="false" show-icon />
      <template v-if="lastTestResult">
        <el-descriptions :border="true" :column="3" size="small">
          <el-descriptions-item label="Datasource">{{ lastTestResult.datasourceId }}</el-descriptions-item>
          <el-descriptions-item label="Status">
            <el-tag :type="lastTestResult.status === 'ok' ? 'success' : 'danger'" size="small">{{ lastTestResult.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Message">{{ lastTestResult.message }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template v-if="lastPreviewResult">
        <p style="margin: 12px 0 8px"><strong>Preview:</strong> {{ lastPreviewResult.datasourceId }} ({{ lastPreviewResult.datasourceType }})</p>
        <el-table v-if="lastPreviewResult.rows.length" :data="lastPreviewResult.rows" size="small" :border="true">
          <el-table-column v-for="col in previewColumns" :key="col" :prop="col" :label="col" />
        </el-table>
        <el-empty v-else description="No rows returned" :image-size="60" />
      </template>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="Create Datasource" width="560px">
      <el-form :model="createForm" label-position="top" @submit.prevent="submitCreateDatasource">
        <el-form-item label="Project ID" required>
          <el-input v-model="createForm.projectId" name="projectId" placeholder="project-1" />
        </el-form-item>
        <el-form-item label="Name" required>
          <el-input v-model="createForm.name" name="name" placeholder="main-db" />
        </el-form-item>
        <el-form-item label="Type">
          <el-select v-model="createForm.type" style="width: 100%">
            <el-option label="PostgreSQL" value="postgresql" />
            <el-option label="Redis" value="redis" />
            <el-option label="MongoDB" value="mongodb" />
          </el-select>
        </el-form-item>
        <el-form-item label="Config (key/value)">
          <div style="width: 100%">
            <div v-for="(pair, index) in createForm.configPairs" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px">
              <el-input v-model="pair.key" placeholder="key" />
              <el-input v-model="pair.value" placeholder="value" />
              <el-button :disabled="creating" @click="removeConfigPair(index)">Remove</el-button>
            </div>
            <el-button :disabled="creating" @click="addConfigPair">Add Config</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button :disabled="creating" @click="closeCreateDialog">Cancel</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreateDatasource">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createDatasource,
  listDatasources,
  previewDatasource,
  testDatasource,
  type Datasource,
  type DatasourcePreviewResult,
  type DatasourceTestResult,
} from '../api/datasources'

type ConfigPair = { key: string; value: string }

const datasources = ref<Datasource[]>([])
const selectedDatasourceId = ref('')
const loading = ref(false)
const creating = ref(false)
const testingId = ref('')
const previewingId = ref('')
const projectFilter = ref('')
const createDialogVisible = ref(false)
const actionError = ref('')
const lastTestResult = ref<DatasourceTestResult | null>(null)
const lastPreviewResult = ref<DatasourcePreviewResult | null>(null)

const createForm = reactive({
  projectId: '',
  name: '',
  type: 'postgresql' as 'postgresql' | 'redis' | 'mongodb',
  configPairs: [{ key: '', value: '' }] as ConfigPair[],
})

const previewColumns = computed(() => {
  if (!lastPreviewResult.value || lastPreviewResult.value.rows.length === 0) return []
  return Object.keys(lastPreviewResult.value.rows[0])
})

function dsTypeTag(type: string) {
  const map: Record<string, string> = { postgresql: '', mongodb: 'success', redis: 'warning' }
  return map[type] || 'info'
}

function configEntries(config: Record<string, string>) {
  return Object.entries(config).map(([key, value]) => ({ key, value }))
}

function resetCreateForm() {
  createForm.projectId = ''
  createForm.name = ''
  createForm.type = 'postgresql'
  createForm.configPairs = [{ key: '', value: '' }]
}

function openCreateDialog() {
  resetCreateForm()
  createDialogVisible.value = true
}

function closeCreateDialog() {
  createDialogVisible.value = false
}

function addConfigPair() {
  createForm.configPairs.push({ key: '', value: '' })
}

function removeConfigPair(index: number) {
  createForm.configPairs.splice(index, 1)
  if (createForm.configPairs.length === 0) {
    createForm.configPairs.push({ key: '', value: '' })
  }
}

function toConfigMap(pairs: ConfigPair[]) {
  const result: Record<string, string> = {}
  for (const pair of pairs) {
    const key = pair.key.trim()
    if (!key) continue
    result[key] = pair.value.trim()
  }
  return result
}

async function loadDatasources() {
  loading.value = true
  try {
    const projectId = projectFilter.value.trim() || undefined
    datasources.value = await listDatasources(projectId)
    if (!datasources.value.some((item) => item.id === selectedDatasourceId.value)) {
      selectedDatasourceId.value = datasources.value[0]?.id ?? ''
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load datasources')
  } finally {
    loading.value = false
  }
}

async function submitCreateDatasource() {
  creating.value = true
  try {
    const config = toConfigMap(createForm.configPairs)
    await createDatasource({
      projectId: createForm.projectId.trim(),
      name: createForm.name.trim(),
      type: createForm.type,
      ...(Object.keys(config).length > 0 ? { config } : {}),
    })
    ElMessage.success('Datasource created')
    await loadDatasources()
    closeCreateDialog()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create datasource')
  } finally {
    creating.value = false
  }
}

async function runTest(id: string) {
  testingId.value = id
  selectedDatasourceId.value = id
  actionError.value = ''
  try {
    lastTestResult.value = await testDatasource(id)
    lastPreviewResult.value = null
  } catch (err) {
    lastTestResult.value = null
    actionError.value = err instanceof Error ? err.message : 'failed to test datasource'
  } finally {
    testingId.value = ''
  }
}

async function runPreview(id: string) {
  previewingId.value = id
  selectedDatasourceId.value = id
  actionError.value = ''
  try {
    lastPreviewResult.value = await previewDatasource(id)
    lastTestResult.value = null
  } catch (err) {
    lastPreviewResult.value = null
    actionError.value = err instanceof Error ? err.message : 'failed to preview datasource'
  } finally {
    previewingId.value = ''
  }
}

onMounted(() => {
  void loadDatasources()
})
</script>
