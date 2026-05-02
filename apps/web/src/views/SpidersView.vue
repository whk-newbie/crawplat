<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Spiders</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Register Docker-based Go or Python spiders and inspect project-scoped spider definitions.
        </p>
      </div>
      <el-button type="primary" @click="createDialogVisible = true">Create Spider</el-button>
    </div>

    <el-card style="margin-bottom: 16px">
      <template #header>Project Spiders</template>
      <div style="display: flex; gap: 12px; margin-bottom: 12px">
        <el-input v-model="projectFilter" placeholder="project id" style="width: 280px" />
        <el-button :loading="loading" @click="loadSpiders">Load Spiders</el-button>
      </div>
      <el-table v-loading="loading" :data="spiders" stripe>
        <el-table-column prop="name" label="Name" />
        <el-table-column prop="language" label="Language">
          <template #default="{ row }">
            <el-tag :type="row.language === 'go' ? 'success' : 'warning'" size="small">{{ row.language }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="runtime" label="Runtime" />
        <el-table-column prop="image" label="Image">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.image }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="140">
          <template #default="{ row }">
            <el-button size="small" @click="openVersions(row)">Versions</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty v-if="!loading" description="No spiders loaded yet" />
        </template>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="Create Spider" width="500px">
      <el-form :model="form" label-position="top" @submit.prevent="submit">
        <el-form-item label="Project ID" required>
          <el-input v-model="form.projectId" />
        </el-form-item>
        <el-form-item label="Name" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="Language">
          <el-select v-model="form.language" style="width: 100%">
            <el-option label="Go" value="go" />
            <el-option label="Python" value="python" />
          </el-select>
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

    <el-dialog
      v-model="versionsDialogVisible"
      :title="selectedSpider ? `Spider Versions · ${selectedSpider.name}` : 'Spider Versions'"
      width="720px"
    >
      <el-form :inline="true" :model="versionForm" @submit.prevent="submitVersion">
        <el-form-item label="Registry Auth Ref">
          <div style="display: flex; gap: 8px; width: 360px">
            <el-select
              v-model="versionForm.registryAuthRef"
              filterable
              allow-create
              default-first-option
              clearable
              placeholder="optional credential ref (e.g. ghcr-prod)"
              style="width: 260px"
            >
              <el-option v-for="item in registryAuthRefs" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button :loading="loadingRegistryAuthRefs" @click="loadRegistryAuthRefs">Load Refs</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Image">
          <el-input v-model="versionForm.image" placeholder="crawler/go:v2" style="width: 280px" />
        </el-form-item>
        <el-form-item label="Command">
          <el-input v-model="versionForm.command" placeholder="./crawler --fast" style="width: 220px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="versionsSubmitting" @click="submitVersion">Create Version</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="versionsLoading" :data="versions" size="small">
        <el-table-column prop="version" label="Version" width="100" />
        <el-table-column label="Registry Auth Ref" width="180">
          <template #default="{ row }">{{ row.registryAuthRef || '-' }}</template>
        </el-table-column>
        <el-table-column prop="image" label="Image" min-width="220" />
        <el-table-column label="Command" min-width="220">
          <template #default="{ row }">{{ Array.isArray(row.command) ? row.command.join(' ') : '' }}</template>
        </el-table-column>
        <el-table-column prop="createdAt" label="Created At" width="180" />
      </el-table>
      <template #footer>
        <el-button @click="versionsDialogVisible = false">Close</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createSpider,
  createSpiderVersion,
  listRegistryAuthRefs,
  listSpiderVersions,
  listSpiders,
  type Spider,
  type SpiderVersion,
} from '../api/spiders'

const form = reactive({
  projectId: 'project-1',
  name: '',
  language: 'go' as 'go' | 'python',
  image: '',
  command: '',
})

const projectFilter = ref('project-1')
const spiders = ref<Spider[]>([])
const versions = ref<SpiderVersion[]>([])
const createDialogVisible = ref(false)
const versionsDialogVisible = ref(false)
const submitting = ref(false)
const loading = ref(false)
const versionsLoading = ref(false)
const versionsSubmitting = ref(false)
const loadingRegistryAuthRefs = ref(false)
const registryAuthRefs = ref<string[]>([])
const selectedSpider = ref<Spider | null>(null)
const versionForm = reactive({
  registryAuthRef: '',
  image: '',
  command: '',
})

function parseCommand(input: string) {
  return input
    .split(' ')
    .map((item) => item.trim())
    .filter(Boolean)
}

async function submit() {
  submitting.value = true
  try {
    const spider = await createSpider({
      projectId: form.projectId,
      name: form.name,
      language: form.language,
      runtime: 'docker',
      image: form.image,
      command: parseCommand(form.command),
    })
    ElMessage.success(`Spider created: ${spider.name}`)
    createDialogVisible.value = false
    await loadSpiders()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create spider')
  } finally {
    submitting.value = false
  }
}

async function loadSpiders() {
  if (!projectFilter.value.trim()) {
    return
  }
  loading.value = true
  try {
    const response = await listSpiders(projectFilter.value.trim())
    spiders.value = response.items
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load spiders')
  } finally {
    loading.value = false
  }
}

async function openVersions(spider: Spider) {
  selectedSpider.value = spider
  versionForm.registryAuthRef = ''
  versionForm.image = spider.image ?? ''
  versionForm.command = Array.isArray(spider.command) ? spider.command.join(' ') : ''
  versionsDialogVisible.value = true
  await loadVersions()
}

async function loadVersions() {
  if (!selectedSpider.value) {
    return
  }
  versionsLoading.value = true
  try {
    versions.value = await listSpiderVersions(selectedSpider.value.id)
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load spider versions')
  } finally {
    versionsLoading.value = false
  }
}

async function loadRegistryAuthRefs() {
  if (!selectedSpider.value) {
    return
  }
  loadingRegistryAuthRefs.value = true
  try {
    const refs = await listRegistryAuthRefs(selectedSpider.value.projectId)
    registryAuthRefs.value = refs
    if (!versionForm.registryAuthRef.trim() && refs.length > 0) {
      versionForm.registryAuthRef = refs[0]
    }
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load registry auth refs')
  } finally {
    loadingRegistryAuthRefs.value = false
  }
}

async function submitVersion() {
  if (!selectedSpider.value) {
    return
  }
  versionsSubmitting.value = true
  try {
    await createSpiderVersion({
      spiderId: selectedSpider.value.id,
      registryAuthRef: versionForm.registryAuthRef.trim() || undefined,
      image: versionForm.image.trim(),
      command: parseCommand(versionForm.command),
    })
    ElMessage.success('Spider version created')
    await loadVersions()
    await loadSpiders()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create spider version')
  } finally {
    versionsSubmitting.value = false
  }
}
</script>
