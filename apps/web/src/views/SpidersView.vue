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
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createSpider, listSpiders, type Spider } from '../api/spiders'

const form = reactive({
  projectId: 'project-1',
  name: '',
  language: 'go' as 'go' | 'python',
  image: '',
  command: '',
})

const projectFilter = ref('project-1')
const spiders = ref<Spider[]>([])
const createDialogVisible = ref(false)
const submitting = ref(false)
const loading = ref(false)

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
</script>
