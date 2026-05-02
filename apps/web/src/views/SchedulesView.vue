<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Schedules</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Create cron-based spider runs and inspect the current retry policy attached to each schedule.
        </p>
      </div>
      <div>
        <el-button :loading="loading" @click="loadSchedules">Refresh</el-button>
        <el-button type="primary" @click="dialogVisible = true">Create Schedule</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="schedules" stripe>
      <el-table-column prop="name" label="Name" />
      <el-table-column label="Spider">
        <template #default="{ row }">
          {{ row.spiderId }}<span v-if="row.spiderVersion"> · v{{ row.spiderVersion }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Registry Auth Ref" width="180">
        <template #default="{ row }">
          {{ row.registryAuthRef || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="cronExpr" label="Cron">
        <template #default="{ row }">
          <el-tag type="info" size="small">{{ row.cronExpr }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="image" label="Image" />
      <el-table-column label="Retry">
        <template #default="{ row }">
          limit: {{ row.retryLimit }} / delay: {{ row.retryDelaySeconds }}s
        </template>
      </el-table-column>
      <el-table-column label="Status" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
            {{ row.enabled ? 'enabled' : 'disabled' }}
          </el-tag>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty v-if="!loading" description="No schedules yet" />
      </template>
    </el-table>

    <el-dialog v-model="dialogVisible" title="Create Schedule" width="520px">
      <el-form :model="form" label-position="top" @submit.prevent="submit">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Project ID" required>
              <el-input v-model="form.projectId" name="projectId" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Spider ID" required>
              <el-input v-model="form.spiderId" name="spiderId" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="Name" required>
          <el-input v-model="form.name" name="name" />
        </el-form-item>
        <el-form-item label="Cron Expression" required>
          <el-input v-model="form.cronExpr" name="cronExpr" placeholder="*/5 * * * *" />
        </el-form-item>
        <el-form-item>
          <el-button :loading="loadingVersions" @click="loadSpiderVersions">Load Spider Versions</el-button>
        </el-form-item>
        <el-form-item label="Spider Version">
          <el-select v-model="form.spiderVersion" clearable placeholder="manual input or pick loaded version" style="width: 100%" @change="applySelectedVersion">
            <el-option
              v-for="item in spiderVersions"
              :key="item.id"
              :label="`v${item.version} · ${item.image}`"
              :value="item.version"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Registry Auth Ref">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input v-model="form.registryAuthRef" placeholder="optional credential ref (e.g. ghcr-prod)" />
            <el-button :loading="loadingRegistryAuthRefs" @click="loadRegistryAuthRefs">Load Registry Refs</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Image">
          <el-input v-model="form.image" name="image" placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item label="Command">
          <el-input v-model="form.command" name="command" placeholder="./go-echo" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Retry Limit">
              <el-input-number v-model="form.retryLimit" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Retry Delay (s)">
              <el-input-number v-model="form.retryDelaySeconds" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="Enabled">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createSchedule, listSchedules, type Schedule } from '../api/schedules'
import { listRegistryAuthRefs, listSpiderVersions, type SpiderVersion } from '../api/spiders'

const schedules = ref<Schedule[]>([])
const loading = ref(true)
const submitting = ref(false)
const dialogVisible = ref(false)
const loadingVersions = ref(false)
const loadingRegistryAuthRefs = ref(false)
const spiderVersions = ref<SpiderVersion[]>([])

const form = reactive({
  projectId: 'project-1',
  spiderId: '',
  spiderVersion: undefined as number | undefined,
  registryAuthRef: '',
  name: '',
  cronExpr: '*/5 * * * *',
  enabled: true,
  image: '',
  command: '',
  retryLimit: 0,
  retryDelaySeconds: 0,
})

function parseCommand(input: string) {
  return input
    .split(' ')
    .map((item) => item.trim())
    .filter(Boolean)
}

function applySelectedVersion() {
  if (form.spiderVersion == null) {
    return
  }
  const selected = spiderVersions.value.find((item) => item.version === form.spiderVersion)
  if (!selected) {
    return
  }
  form.registryAuthRef = selected.registryAuthRef ?? ''
  form.image = selected.image
  form.command = Array.isArray(selected.command) ? selected.command.join(' ') : ''
}

async function loadSchedules() {
  loading.value = true
  try {
    const response = await listSchedules()
    schedules.value = response.items
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load schedules')
  } finally {
    loading.value = false
  }
}

async function submit() {
  submitting.value = true
  try {
    const schedule = await createSchedule({
      projectId: form.projectId,
      spiderId: form.spiderId,
      spiderVersion: form.spiderVersion,
      registryAuthRef: form.registryAuthRef.trim() || undefined,
      name: form.name,
      cronExpr: form.cronExpr,
      enabled: form.enabled,
      image: form.image.trim() || undefined,
      command: parseCommand(form.command),
      retryLimit: form.retryLimit,
      retryDelaySeconds: form.retryDelaySeconds,
    })
    ElMessage.success(`Schedule created: ${schedule.name}`)
    schedules.value = [schedule, ...schedules.value]
    dialogVisible.value = false
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create schedule')
  } finally {
    submitting.value = false
  }
}

async function loadSpiderVersions() {
  const spiderID = form.spiderId.trim()
  if (!spiderID) {
    ElMessage.warning('Spider ID is required')
    return
  }
  loadingVersions.value = true
  try {
    spiderVersions.value = await listSpiderVersions(spiderID)
    if (spiderVersions.value.length > 0) {
      form.spiderVersion = spiderVersions.value[0].version
      applySelectedVersion()
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load spider versions')
  } finally {
    loadingVersions.value = false
  }
}

async function loadRegistryAuthRefs() {
  const projectID = form.projectId.trim()
  if (!projectID) {
    ElMessage.warning('Project ID is required')
    return
  }
  loadingRegistryAuthRefs.value = true
  try {
    const refs = await listRegistryAuthRefs(projectID)
    if (!form.registryAuthRef.trim() && refs.length > 0) {
      form.registryAuthRef = refs[0]
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load registry auth refs')
  } finally {
    loadingRegistryAuthRefs.value = false
  }
}

onMounted(() => {
  void loadSchedules()
})
</script>
