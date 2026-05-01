<template>
  <main class="page">
    <section class="card hero">
      <div>
        <p class="eyebrow">Data access</p>
        <h1>Datasources</h1>
        <p>Manage data connections and run connectivity checks or row preview.</p>
      </div>
      <div class="toolbar">
        <button :disabled="loading" @click="loadDatasources">
          {{ loading ? 'Refreshing...' : 'Refresh' }}
        </button>
        <button @click="openCreateDialog">Create Datasource</button>
      </div>
    </section>

    <section class="card">
      <div class="toolbar wrap">
        <label>
          Project ID
          <input v-model.trim="projectFilter" placeholder="project-1" />
        </label>
        <button :disabled="loading" @click="loadDatasources">Query</button>
      </div>
      <p v-if="loading">Loading datasources...</p>
      <p v-else-if="error" class="error">{{ error }}</p>
      <p v-else-if="datasources.length === 0">No datasources found.</p>
      <ul v-else class="datasource-list">
        <li v-for="datasource in datasources" :key="datasource.id">
          <article :class="['item', { active: datasource.id === selectedDatasourceId }]">
            <header class="item-head">
              <div>
                <strong>{{ datasource.name }}</strong>
                <small>{{ datasource.type }}</small>
              </div>
              <code>{{ datasource.id }}</code>
            </header>
            <p class="meta">project: {{ datasource.projectId }} | readonly: {{ datasource.readonly }}</p>
            <dl class="config-list">
              <div v-for="entry in configEntries(datasource.config)" :key="entry.key">
                <dt>{{ entry.key }}</dt>
                <dd>{{ entry.value }}</dd>
              </div>
              <p v-if="configEntries(datasource.config).length === 0">No config entries.</p>
            </dl>
            <div class="actions">
              <button :disabled="testingId === datasource.id" @click="runTest(datasource.id)">
                {{ testingId === datasource.id ? 'Testing...' : 'Test' }}
              </button>
              <button :disabled="previewingId === datasource.id" @click="runPreview(datasource.id)">
                {{ previewingId === datasource.id ? 'Loading...' : 'Preview' }}
              </button>
            </div>
          </article>
        </li>
      </ul>
    </section>

    <section class="card">
      <h2>Action Result</h2>
      <p v-if="actionError" class="error">{{ actionError }}</p>
      <template v-else>
        <template v-if="lastTestResult">
        <p><strong>Test Result</strong></p>
        <p>datasource: <code>{{ lastTestResult.datasourceId }}</code></p>
        <p>status: {{ lastTestResult.status }}</p>
        <p>message: {{ lastTestResult.message }}</p>
        </template>
        <template v-if="lastPreviewResult">
        <p><strong>Preview Result</strong></p>
        <p>
          datasource: <code>{{ lastPreviewResult.datasourceId }}</code> ({{ lastPreviewResult.datasourceType }})
        </p>
        <p v-if="lastPreviewResult.rows.length === 0">No rows returned.</p>
        <ul v-else class="preview-list">
          <li v-for="(row, rowIndex) in lastPreviewResult.rows" :key="rowIndex">
            <code>{{ JSON.stringify(row) }}</code>
          </li>
        </ul>
        </template>
        <p v-if="!lastTestResult && !lastPreviewResult">Select one datasource and run test/preview.</p>
      </template>
    </section>

    <section v-show="createDialogVisible" class="dialog-backdrop">
      <article class="dialog">
        <h2>Create Datasource</h2>
        <form class="form" @submit.prevent="submitCreateDatasource">
          <label>
            Project ID
            <input v-model="createForm.projectId" name="projectId" required placeholder="project-1" />
          </label>
          <label>
            Name
            <input v-model="createForm.name" name="name" required placeholder="main-db" />
          </label>
          <label>
            Type
            <select v-model="createForm.type" name="type">
              <option value="postgresql">postgresql</option>
              <option value="redis">redis</option>
              <option value="mongodb">mongodb</option>
            </select>
          </label>
          <section class="config-editor">
            <header>
              <h3>Config (key/value)</h3>
              <button :disabled="creating" type="button" @click="addConfigPair">Add</button>
            </header>
            <div v-for="(pair, index) in createForm.configPairs" :key="index" class="config-row">
              <input
                v-model="pair.key"
                :name="`config-key-${index}`"
                placeholder="key"
              />
              <input
                v-model="pair.value"
                :name="`config-value-${index}`"
                placeholder="value"
              />
              <button :disabled="creating" type="button" @click="removeConfigPair(index)">
                Remove
              </button>
            </div>
          </section>
          <div class="actions">
            <button :disabled="creating" type="button" @click="closeCreateDialog">Cancel</button>
            <button :disabled="creating" type="submit">
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
        <p v-if="createError" class="error">{{ createError }}</p>
      </article>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
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
const error = ref('')
const createError = ref('')
const actionError = ref('')
const projectFilter = ref('')
const createDialogVisible = ref(false)
const lastTestResult = ref<DatasourceTestResult | null>(null)
const lastPreviewResult = ref<DatasourcePreviewResult | null>(null)

const createForm = reactive({
  projectId: '',
  name: '',
  type: 'postgresql' as 'postgresql' | 'redis' | 'mongodb',
  configPairs: [{ key: '', value: '' }] as ConfigPair[],
})

function configEntries(config: Record<string, string>) {
  return Object.entries(config).map(([key, value]) => ({ key, value }))
}

function resetCreateForm() {
  createForm.projectId = ''
  createForm.name = ''
  createForm.type = 'postgresql'
  createForm.configPairs = [{ key: '', value: '' }]
  createError.value = ''
}

function openCreateDialog() {
  resetCreateForm()
  createDialogVisible.value = true
}

function closeCreateDialog() {
  createDialogVisible.value = false
  createError.value = ''
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
    if (!key) {
      continue
    }
    result[key] = pair.value.trim()
  }
  return result
}

async function loadDatasources() {
  loading.value = true
  error.value = ''
  try {
    const projectId = projectFilter.value.trim() || undefined
    datasources.value = await listDatasources(projectId)
    if (!datasources.value.some((item) => item.id === selectedDatasourceId.value)) {
      selectedDatasourceId.value = datasources.value[0]?.id ?? ''
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to load datasources'
  } finally {
    loading.value = false
  }
}

async function submitCreateDatasource() {
  creating.value = true
  createError.value = ''
  try {
    const config = toConfigMap(createForm.configPairs)
    const payload = {
      projectId: createForm.projectId.trim(),
      name: createForm.name.trim(),
      type: createForm.type,
      ...(Object.keys(config).length > 0 ? { config } : {}),
    }
    await createDatasource({
      ...payload,
    })
    await loadDatasources()
    closeCreateDialog()
  } catch (err) {
    createError.value = err instanceof Error ? err.message : 'failed to create datasource'
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
  align-items: flex-start;
  gap: 1rem;
}

.eyebrow {
  margin: 0 0 0.25rem;
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #57606a;
}

.toolbar {
  display: flex;
  gap: 0.75rem;
  align-items: end;
}

.toolbar.wrap {
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
}

.toolbar label {
  display: grid;
  gap: 0.25rem;
  font-size: 0.85rem;
}

.datasource-list {
  display: grid;
  gap: 0.75rem;
  list-style: none;
  margin: 0;
  padding: 0;
}

.item {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  background: #f6f8fa;
  padding: 0.75rem;
  display: grid;
  gap: 0.5rem;
}

.item.active {
  border-color: #0969da;
  background: #ddf4ff;
}

.item-head {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: baseline;
}

.item-head div {
  display: grid;
  gap: 0.2rem;
}

.meta {
  margin: 0;
  color: #57606a;
}

.config-list {
  display: grid;
  gap: 0.35rem;
  margin: 0;
}

.config-list div {
  display: grid;
  grid-template-columns: 8rem minmax(0, 1fr);
  gap: 0.5rem;
}

.config-list dt {
  font-weight: 600;
}

.config-list dd {
  margin: 0;
  word-break: break-all;
}

.preview-list {
  display: grid;
  gap: 0.5rem;
}

.preview-list code {
  display: block;
  padding: 0.45rem 0.6rem;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  background: #f6f8fa;
}

.dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgb(12 17 29 / 45%);
  display: grid;
  place-items: center;
  padding: 1rem;
}

.dialog {
  width: min(40rem, 100%);
  border: 1px solid #d0d7de;
  border-radius: 8px;
  background: #fff;
  padding: 1rem;
  display: grid;
  gap: 0.75rem;
}

.form {
  display: grid;
  gap: 0.75rem;
}

.config-editor {
  display: grid;
  gap: 0.5rem;
}

.config-editor header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-editor h3 {
  margin: 0;
  font-size: 1rem;
}

.config-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  gap: 0.5rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.error {
  color: #b42318;
}

@media (max-width: 860px) {
  .hero {
    flex-direction: column;
  }

  .config-list div {
    grid-template-columns: 1fr;
  }

  .config-row {
    grid-template-columns: 1fr;
  }
}
</style>
