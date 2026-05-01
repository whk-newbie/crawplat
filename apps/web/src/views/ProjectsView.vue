<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Projects</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Manage project namespaces for spiders, schedules, and executions.
        </p>
      </div>
      <div>
        <el-button :loading="loading" @click="loadProjects">Refresh</el-button>
        <el-button type="primary" @click="openCreateDialog">Create Project</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="projects" stripe>
      <el-table-column prop="name" label="Name" />
      <el-table-column prop="code" label="Code" />
      <el-table-column prop="id" label="ID">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.id }}</el-tag>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty v-if="!loading" description="No projects yet" />
      </template>
    </el-table>

    <el-dialog v-model="createDialogVisible" title="Create Project" width="480px">
      <el-form :model="createForm" label-position="top" @submit.prevent="submitCreateProject">
        <el-form-item label="Code" required>
          <el-input v-model="createForm.code" name="code" placeholder="core-crawlers" />
        </el-form-item>
        <el-form-item label="Name" required>
          <el-input v-model="createForm.name" name="name" placeholder="Core Crawlers" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button :disabled="creating" @click="closeCreateDialog">Cancel</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreateProject">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createProject, listProjects, type Project } from '../api/projects'

const projects = ref<Project[]>([])
const loading = ref(true)
const creating = ref(false)
const createDialogVisible = ref(false)

const createForm = reactive({
  code: '',
  name: '',
})

function resetCreateForm() {
  createForm.code = ''
  createForm.name = ''
}

function openCreateDialog() {
  resetCreateForm()
  createDialogVisible.value = true
}

function closeCreateDialog() {
  createDialogVisible.value = false
}

async function loadProjects() {
  loading.value = true
  try {
    const response = await listProjects()
    projects.value = response.items
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to load projects')
  } finally {
    loading.value = false
  }
}

async function submitCreateProject() {
  creating.value = true
  try {
    await createProject({
      code: createForm.code,
      name: createForm.name,
    })
    ElMessage.success('Project created')
    await loadProjects()
    closeCreateDialog()
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create project')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  void loadProjects()
})
</script>
