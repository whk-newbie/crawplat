<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Project } from '../api/projects'
import { createProject, listProjects } from '../api/projects'
import { ApiError } from '../api/client'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()

const projects = ref<Project[]>([])
const loading = ref(false)
const loadError = ref('')
const showCreateDialog = ref(false)

const createForm = reactive({ code: '', name: '' })
const creating = ref(false)

async function loadProjects() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listProjects()
    projects.value = res.items ?? []
  } catch (err) {
    loadError.value = localeStore.t('pages.projects.errors.loadFailed')
    projects.value = []
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  createForm.code = ''
  createForm.name = ''
  showCreateDialog.value = true
}

async function handleCreate() {
  if (!createForm.code.trim()) {
    ElMessage.warning(localeStore.t('pages.projects.errors.codeRequired'))
    return
  }
  if (!createForm.name.trim()) {
    ElMessage.warning(localeStore.t('pages.projects.errors.nameRequired'))
    return
  }

  creating.value = true
  try {
    await createProject({
      code: createForm.code.trim(),
      name: createForm.name.trim(),
    })
    ElMessage.success(localeStore.t('pages.projects.createSuccess'))
    showCreateDialog.value = false
    await loadProjects()
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    } else {
      ElMessage.error(localeStore.t('pages.projects.errors.createFailed'))
    }
  } finally {
    creating.value = false
  }
}

onMounted(loadProjects)
</script>

<template>
  <main class="projects-page">
    <div class="page-header">
      <h1>{{ localeStore.t('pages.projects.title') }}</h1>
      <el-button type="primary" @click="openCreateDialog">
        {{ localeStore.t('pages.projects.create') }}
      </el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="projects"
      stripe
      :empty-text="localeStore.t('pages.projects.empty')"
    >
      <el-table-column
        prop="code"
        :label="localeStore.t('pages.projects.code')"
        min-width="160"
      />
      <el-table-column
        prop="name"
        :label="localeStore.t('pages.projects.name')"
        min-width="200"
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
      :title="localeStore.t('pages.projects.create')"
    >
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="localeStore.t('pages.projects.code')">
          <el-input v-model="createForm.code" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.projects.name')">
          <el-input v-model="createForm.name" />
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
.projects-page {
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
