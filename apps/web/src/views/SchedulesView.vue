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
        <el-form-item label="Image" required>
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

const schedules = ref<Schedule[]>([])
const loading = ref(true)
const submitting = ref(false)
const dialogVisible = ref(false)

const form = reactive({
  projectId: 'project-1',
  spiderId: '',
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
      name: form.name,
      cronExpr: form.cronExpr,
      enabled: form.enabled,
      image: form.image,
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

onMounted(() => {
  void loadSchedules()
})
</script>
