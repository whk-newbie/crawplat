<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Executions</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Create a manual execution and jump into its detail page, or inspect an existing execution by ID.
        </p>
      </div>
      <el-button type="primary" @click="createDialogVisible = true">Create Execution</el-button>
    </div>

    <el-card>
      <template #header>Open Execution Detail</template>
      <div style="display: flex; gap: 12px">
        <el-input v-model="lookupId" placeholder="execution id" style="width: 360px" />
        <el-button @click="openExecution">Open</el-button>
      </div>
    </el-card>

    <el-card style="margin-top: 16px">
      <template #header>
        <div style="display: flex; align-items: center; justify-content: space-between">
          <span>Recent Executions</span>
          <el-button :loading="loadingList" @click="loadExecutions">Refresh</el-button>
        </div>
      </template>
      <div style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px">
        <el-input v-model="executionSpiderId" clearable placeholder="spider id" style="width: 180px" />
        <el-select v-model="executionStatus" clearable placeholder="status" style="width: 180px">
          <el-option label="pending" value="pending" />
          <el-option label="running" value="running" />
          <el-option label="succeeded" value="succeeded" />
          <el-option label="failed" value="failed" />
        </el-select>
        <el-select v-model="executionTriggerSource" clearable placeholder="trigger" style="width: 180px">
          <el-option label="manual" value="manual" />
          <el-option label="scheduled" value="scheduled" />
          <el-option label="retry" value="retry" />
        </el-select>
        <el-date-picker
          v-model="executionTimeRange"
          type="datetimerange"
          start-placeholder="from"
          end-placeholder="to"
          style="width: 380px"
        />
        <el-button @click="applyExecutionFilters">Apply Filters</el-button>
        <el-button @click="resetExecutionFilters">Reset</el-button>
      </div>
      <el-table v-loading="loadingList" :data="executions" stripe>
        <el-table-column prop="id" label="Execution ID" min-width="220">
          <template #default="{ row }">
            <el-button link type="primary" @click="openExecutionById(row.id)">{{ row.id }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="spiderId" label="Spider" min-width="140" />
        <el-table-column prop="status" label="Status" width="120" />
        <el-table-column prop="triggerSource" label="Trigger" width="120" />
        <el-table-column prop="createdAt" label="Created At" min-width="200" />
        <template #empty>
          <el-empty v-if="!loadingList" description="No executions" />
        </template>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 12px">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          layout="prev, pager, next, total"
          :total="total"
          :page-sizes="[20]"
          @current-change="loadExecutions"
        />
      </div>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="Create Execution" width="500px">
      <el-form :model="form" label-position="top" @submit.prevent="submit">
        <el-form-item label="Project ID" required>
          <el-input v-model="form.projectId" />
        </el-form-item>
        <el-form-item label="Spider ID" required>
          <el-input v-model="form.spiderId" />
        </el-form-item>
        <el-form-item>
          <el-button :loading="loadingVersions" @click="loadSpiderVersions">Load Spider Versions</el-button>
        </el-form-item>
        <el-form-item label="Spider Version">
          <el-select v-model="form.spiderVersion" clearable placeholder="latest from loaded list" style="width: 100%" @change="applySelectedVersion">
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
            <el-select
              v-model="form.registryAuthRef"
              filterable
              allow-create
              default-first-option
              clearable
              placeholder="optional credential ref (e.g. ghcr-prod)"
              style="width: 100%"
            >
              <el-option v-for="item in registryAuthRefs" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button :loading="loadingRegistryAuthRefs" @click="loadRegistryAuthRefs">Load Registry Refs</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Image" required>
          <el-input v-model="form.image" placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item label="Command">
          <el-input v-model="form.command" placeholder="./go-echo" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createExecution, listExecutions, type Execution } from '../api/executions'
import { listRegistryAuthRefs, listSpiderVersions, type SpiderVersion } from '../api/spiders'

const router = useRouter()

const form = reactive({
  projectId: 'project-1',
  spiderId: '',
  spiderVersion: undefined as number | undefined,
  registryAuthRef: '',
  image: '',
  command: '',
})

const lookupId = ref('')
const createDialogVisible = ref(false)
const submitting = ref(false)
const loadingVersions = ref(false)
const loadingRegistryAuthRefs = ref(false)
const spiderVersions = ref<SpiderVersion[]>([])
const registryAuthRefs = ref<string[]>([])
const executions = ref<Execution[]>([])
const loadingList = ref(false)
const total = ref(0)
const pageSize = ref(20)
const currentPage = ref(1)
const executionSpiderId = ref('')
const executionStatus = ref('')
const executionTriggerSource = ref('')
const executionTimeRange = ref<[Date, Date] | null>(null)

function parseCommand(input: string) {
  return input
    .split(' ')
    .map((item) => item.trim())
    .filter(Boolean)
}

async function submit() {
  submitting.value = true
  try {
    const execution = await createExecution({
      projectId: form.projectId,
      spiderId: form.spiderId,
      spiderVersion: form.spiderVersion,
      registryAuthRef: form.registryAuthRef.trim() || undefined,
      image: form.image,
      command: parseCommand(form.command),
    })
    ElMessage.success('Execution created')
    createDialogVisible.value = false
    lookupId.value = execution.id
    await router.push(`/executions/${execution.id}`)
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create execution')
  } finally {
    submitting.value = false
  }
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
    registryAuthRefs.value = refs
    if (!form.registryAuthRef.trim() && refs.length > 0) {
      form.registryAuthRef = refs[0]
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load registry auth refs')
  } finally {
    loadingRegistryAuthRefs.value = false
  }
}

async function openExecution() {
  if (!lookupId.value.trim()) {
    return
  }
  await router.push(`/executions/${lookupId.value.trim()}`)
}

async function openExecutionById(executionID: string) {
  await router.push(`/executions/${executionID}`)
}

async function loadExecutions() {
  const projectID = form.projectId.trim()
  if (!projectID) {
    ElMessage.warning('Project ID is required')
    return
  }
  let executionFrom: string | undefined
  let executionTo: string | undefined
  if (executionTimeRange.value && executionTimeRange.value.length === 2) {
    executionFrom = executionTimeRange.value[0].toISOString()
    executionTo = executionTimeRange.value[1].toISOString()
  }
  loadingList.value = true
  try {
    const response = await listExecutions({
      projectId: projectID,
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      spiderId: executionSpiderId.value || undefined,
      executionStatus: executionStatus.value || undefined,
      executionTriggerSource: executionTriggerSource.value || undefined,
      executionFrom,
      executionTo,
    })
    executions.value = response.items
    total.value = response.total
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load executions')
  } finally {
    loadingList.value = false
  }
}

async function applyExecutionFilters() {
  currentPage.value = 1
  await loadExecutions()
}

async function resetExecutionFilters() {
  executionSpiderId.value = ''
  executionStatus.value = ''
  executionTriggerSource.value = ''
  executionTimeRange.value = null
  currentPage.value = 1
  await loadExecutions()
}

onMounted(() => {
  void loadExecutions()
})
</script>
