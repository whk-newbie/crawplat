<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Schedule } from '../api/schedules'
import { createSchedule, listSchedules } from '../api/schedules'
import { listProjects, type Project } from '../api/projects'
import { listSpiders, type Spider } from '../api/spiders'
import { ApiError } from '../api/client'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()

const schedules = ref<Schedule[]>([])
const loading = ref(false)
const loadError = ref('')
const showCreateDialog = ref(false)

const projects = ref<Project[]>([])
const spiders = ref<Spider[]>([])
const selectedProjectId = ref('')

const createForm = reactive({
  spiderId: '',
  name: '',
  cronExpr: '',
  enabled: true,
  image: '',
  command: '',
  retryLimit: 3,
  retryDelaySeconds: 60,
})
const creating = ref(false)

async function loadData() {
  try {
    const projRes = await listProjects()
    projects.value = projRes.items ?? []
    if (projects.value.length > 0) {
      selectedProjectId.value = projects.value[0].id
    }
  } catch { /* ignore */ }
  await loadSchedules()
}

async function loadSchedules() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listSchedules()
    schedules.value = res.items ?? []
  } catch {
    loadError.value = localeStore.t('pages.schedules.errors.loadFailed')
    schedules.value = []
  } finally {
    loading.value = false
  }
}

async function onProjectChange(projectId: string) {
  selectedProjectId.value = projectId
  try {
    spiders.value = await listSpiders(projectId)
  } catch {
    spiders.value = []
  }
}

function openCreateDialog() {
  createForm.spiderId = ''
  createForm.name = ''
  createForm.cronExpr = ''
  createForm.enabled = true
  createForm.image = ''
  createForm.command = ''
  createForm.retryLimit = 3
  createForm.retryDelaySeconds = 60
  showCreateDialog.value = true
}

async function handleCreate() {
  if (!createForm.name.trim()) {
    ElMessage.warning(localeStore.t('pages.schedules.errors.nameRequired'))
    return
  }
  if (!createForm.cronExpr.trim()) {
    ElMessage.warning(localeStore.t('pages.schedules.errors.cronRequired'))
    return
  }
  if (!createForm.spiderId) {
    ElMessage.warning(localeStore.t('pages.schedules.errors.spiderIdRequired'))
    return
  }

  creating.value = true
  try {
    await createSchedule({
      projectId: selectedProjectId.value,
      spiderId: createForm.spiderId,
      name: createForm.name.trim(),
      cronExpr: createForm.cronExpr.trim(),
      enabled: createForm.enabled,
      image: createForm.image.trim() || undefined,
      command: createForm.command
        .split(' ')
        .map((s) => s.trim())
        .filter(Boolean),
      retryLimit: createForm.retryLimit,
      retryDelaySeconds: createForm.retryDelaySeconds,
    })
    ElMessage.success(localeStore.t('pages.schedules.createSuccess'))
    showCreateDialog.value = false
    await loadSchedules()
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    } else {
      ElMessage.error(localeStore.t('pages.schedules.errors.createFailed'))
    }
  } finally {
    creating.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <main class="schedules-page">
    <div class="page-header">
      <h1>{{ localeStore.t('pages.schedules.title') }}</h1>
      <el-button type="primary" @click="openCreateDialog">
        {{ localeStore.t('pages.schedules.create') }}
      </el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="schedules"
      stripe
      :empty-text="localeStore.t('pages.schedules.empty')"
    >
      <el-table-column prop="name" :label="localeStore.t('pages.schedules.name')" min-width="140" />
      <el-table-column prop="cronExpr" :label="localeStore.t('pages.schedules.cronExpr')" min-width="140" />
      <el-table-column :label="localeStore.t('pages.schedules.enabled')" width="80">
        <template #default="scope">
          <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
            {{ scope.row.enabled ? 'ON' : 'OFF' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="retryLimit" :label="localeStore.t('pages.schedules.retryLimit')" width="100" />
      <el-table-column
        prop="retryDelaySeconds"
        :label="localeStore.t('pages.schedules.retryDelaySeconds')"
        width="120"
      />
    </el-table>

    <el-alert
      v-if="loadError"
      :title="loadError"
      type="error"
      show-icon
      class="error-alert"
    />

    <!-- Create Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="localeStore.t('pages.schedules.create')"
    >
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="localeStore.t('pages.schedules.projectId')">
          <el-select v-model="selectedProjectId" @change="onProjectChange">
            <el-option
              v-for="p in projects"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.spiderId')">
          <el-select v-model="createForm.spiderId">
            <el-option
              v-for="s in spiders"
              :key="s.id"
              :label="s.name"
              :value="s.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.name')">
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.cronExpr')">
          <el-input v-model="createForm.cronExpr" :placeholder="localeStore.t('pages.schedules.cronHint')" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.enabled')">
          <el-switch v-model="createForm.enabled" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.image')">
          <el-input v-model="createForm.image" placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.command')">
          <el-input v-model="createForm.command" placeholder="./go-echo" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.retryLimit')">
          <el-input-number v-model="createForm.retryLimit" :min="0" :max="10" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.schedules.retryDelaySeconds')">
          <el-input-number v-model="createForm.retryDelaySeconds" :min="0" :max="3600" />
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
.schedules-page {
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
  margin-top: 1rem;
}
</style>
