<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Datasource, DatasourcePreviewResult, DatasourceTestResult } from '../api/datasources'
import { createDatasource, listDatasources, previewDatasource, testDatasource } from '../api/datasources'
import { listProjects, type Project } from '../api/projects'
import { ApiError } from '../api/client'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()

const projects = ref<Project[]>([])
const selectedProjectId = ref('')
const datasources = ref<Datasource[]>([])
const loading = ref(false)
const loadError = ref('')

const showCreateDialog = ref(false)
const createForm = reactive({
  name: '',
  type: 'postgresql' as string,
  configKey: '',
  configValue: '',
})
const creating = ref(false)

const datasourceTypes = ['postgresql', 'mongodb', 'redis']

const testResult = ref<DatasourceTestResult | null>(null)
const previewResult = ref<DatasourcePreviewResult | null>(null)
const testing = ref(false)
const previewing = ref(false)

async function loadData() {
  try {
    const res = await listProjects()
    projects.value = res.items ?? []
    if (projects.value.length > 0) {
      selectedProjectId.value = projects.value[0].id
    }
  } catch { /* ignore */ }
  await loadDatasources()
}

async function loadDatasources() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listDatasources(selectedProjectId.value || undefined)
    datasources.value = res.items ?? []
  } catch {
    loadError.value = localeStore.t('pages.datasources.errors.loadFailed')
    datasources.value = []
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  createForm.name = ''
  createForm.type = 'postgresql'
  createForm.configKey = ''
  createForm.configValue = ''
  showCreateDialog.value = true
}

async function handleCreate() {
  if (!createForm.name.trim()) {
    ElMessage.warning(localeStore.t('pages.datasources.errors.nameRequired'))
    return
  }

  creating.value = true
  try {
    await createDatasource({
      projectId: selectedProjectId.value,
      name: createForm.name.trim(),
      type: createForm.type as 'mongodb' | 'redis' | 'postgresql',
      config: createForm.configKey.trim()
        ? { [createForm.configKey.trim()]: createForm.configValue.trim() }
        : undefined,
    })
    ElMessage.success(localeStore.t('pages.datasources.createSuccess'))
    showCreateDialog.value = false
    await loadDatasources()
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    } else {
      ElMessage.error(localeStore.t('pages.datasources.errors.createFailed'))
    }
  } finally {
    creating.value = false
  }
}

async function handleTest(ds: Datasource) {
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testDatasource(ds.id)
  } catch (err) {
    ElMessage.error(localeStore.t('pages.datasources.errors.testFailed'))
  } finally {
    testing.value = false
  }
}

async function handlePreview(ds: Datasource) {
  previewing.value = true
  previewResult.value = null
  try {
    previewResult.value = await previewDatasource(ds.id)
  } catch {
    previewResult.value = null
  } finally {
    previewing.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <main class="datasources-page">
    <div class="page-header">
      <h1>{{ localeStore.t('pages.datasources.title') }}</h1>
      <el-button type="primary" @click="openCreateDialog">
        {{ localeStore.t('pages.datasources.create') }}
      </el-button>
    </div>

    <div class="project-selector">
      <el-select
        v-model="selectedProjectId"
        :placeholder="localeStore.t('pages.spiders.projectId')"
        @change="loadDatasources"
      >
        <el-option
          v-for="p in projects"
          :key="p.id"
          :label="p.name"
          :value="p.id"
        />
      </el-select>
    </div>

    <el-table
      v-loading="loading"
      :data="datasources"
      stripe
      :empty-text="localeStore.t('pages.datasources.empty')"
    >
      <el-table-column prop="name" :label="localeStore.t('pages.datasources.name')" min-width="140" />
      <el-table-column prop="type" :label="localeStore.t('pages.datasources.type')" width="120" />
      <el-table-column :label="localeStore.t('pages.datasources.readonly')" width="80">
        <template #default="scope">
          <el-tag :type="scope.row.readonly ? 'warning' : 'success'" size="small">
            {{ scope.row.readonly ? 'Yes' : 'No' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="localeStore.t('common.actions.create')" width="180">
        <template #default="scope">
          <el-button size="small" :loading="testing" @click="handleTest(scope.row)">
            {{ localeStore.t('pages.datasources.test') }}
          </el-button>
          <el-button size="small" :loading="previewing" @click="handlePreview(scope.row)">
            {{ localeStore.t('pages.datasources.preview') }}
          </el-button>
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

    <!-- Test Result -->
    <el-card v-if="testResult" class="result-card">
      <template #header>
        <h3>{{ localeStore.t('pages.datasources.testResult') }}</h3>
      </template>
      <p>Status: <el-tag :type="testResult.status === 'ok' ? 'success' : 'danger'" size="small">{{ testResult.status }}</el-tag></p>
      <p v-if="testResult.message">{{ testResult.message }}</p>
    </el-card>

    <!-- Preview Result -->
    <el-card v-if="previewResult" class="result-card">
      <template #header>
        <h3>{{ localeStore.t('pages.datasources.previewResult') }}</h3>
      </template>
      <el-table v-if="previewResult.rows?.length" :data="previewResult.rows" size="small" max-height="300">
        <el-table-column
          v-for="col in Object.keys(previewResult.rows[0] || {})"
          :key="col"
          :prop="col"
          :label="col"
          min-width="100"
        />
      </el-table>
      <el-empty v-else :description="localeStore.t('common.status.empty')" />
    </el-card>

    <!-- Create Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="localeStore.t('pages.datasources.create')"
    >
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="localeStore.t('pages.datasources.name')">
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.datasources.type')">
          <el-select v-model="createForm.type">
            <el-option
              v-for="t in datasourceTypes"
              :key="t"
              :label="t"
              :value="t"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.datasources.config')">
          <el-input v-model="createForm.configKey" placeholder="key (e.g. schema)" />
          <el-input v-model="createForm.configValue" placeholder="value" style="margin-top: 0.5rem" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">
          {{ localeStore.t('common.actions.cancel') }}
        </el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">
          {{ localeStore.t('common.actions.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </main>
</template>

<style scoped>
.datasources-page {
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

.project-selector {
  margin-bottom: 1rem;
}

.error-alert {
  margin-top: 1rem;
}

.result-card {
  margin-top: 1rem;
}
</style>
