<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Spider } from '../api/spiders'
import {
  createSpider,
  createSpiderVersion,
  listRegistryAuthRefs,
  listSpiders,
  listSpiderVersions,
  type RegistryAuthRef,
  type SpiderVersion,
} from '../api/spiders'
import { ApiError } from '../api/client'
import { listProjects, type Project } from '../api/projects'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()

const projects = ref<Project[]>([])
const selectedProjectId = ref('')
const spiders = ref<Spider[]>([])
const loading = ref(false)
const loadError = ref('')

const showCreateDialog = ref(false)
const createForm = reactive({
  name: '',
  language: 'go' as 'go' | 'python',
  image: '',
  command: '',
})
const creating = ref(false)

const showVersionsDialog = ref(false)
const selectedSpider = ref<Spider | null>(null)
const versions = ref<SpiderVersion[]>([])
const loadingVersions = ref(false)
const versionForm = reactive({ version: '', image: '' })
const creatingVersion = ref(false)

async function loadProjects() {
  try {
    const res = await listProjects()
    projects.value = res.items ?? []
    if (projects.value.length > 0) {
      selectedProjectId.value = projects.value[0].id
      await loadSpiders()
    }
  } catch {
    // projects not critical for spiders
  }
}

async function loadSpiders() {
  if (!selectedProjectId.value) return
  loading.value = true
  loadError.value = ''
  try {
    spiders.value = await listSpiders(selectedProjectId.value)
  } catch {
    loadError.value = localeStore.t('pages.spiders.errors.loadFailed')
    spiders.value = []
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  createForm.name = ''
  createForm.language = 'go'
  createForm.image = ''
  createForm.command = ''
  showCreateDialog.value = true
}

async function handleCreate() {
  if (!createForm.name.trim()) {
    ElMessage.warning(localeStore.t('pages.spiders.errors.nameRequired'))
    return
  }
  if (!createForm.image.trim()) {
    ElMessage.warning(localeStore.t('pages.spiders.errors.imageRequired'))
    return
  }

  creating.value = true
  try {
    await createSpider({
      projectId: selectedProjectId.value,
      name: createForm.name.trim(),
      language: createForm.language,
      runtime: 'docker',
      image: createForm.image.trim(),
      command: createForm.command
        .split(' ')
        .map((s) => s.trim())
        .filter(Boolean),
    })
    ElMessage.success(localeStore.t('pages.spiders.createSuccess'))
    showCreateDialog.value = false
    await loadSpiders()
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    } else {
      ElMessage.error(localeStore.t('pages.spiders.errors.createFailed'))
    }
  } finally {
    creating.value = false
  }
}

async function openVersionsDialog(spider: Spider) {
  selectedSpider.value = spider
  loadingVersions.value = true
  versionForm.version = ''
  versionForm.image = spider.image ?? ''
  try {
    versions.value = await listSpiderVersions(spider.id)
  } catch {
    versions.value = []
  } finally {
    loadingVersions.value = false
  }
  showVersionsDialog.value = true
}

async function handleCreateVersion() {
  if (!selectedSpider.value || !versionForm.version.trim()) return
  creatingVersion.value = true
  try {
    await createSpiderVersion(selectedSpider.value.id, {
      version: versionForm.version.trim(),
      image: versionForm.image.trim(),
      isCurrent: false,
    })
    ElMessage.success(localeStore.t('pages.spiders.versions.createSuccess'))
    versions.value = await listSpiderVersions(selectedSpider.value.id)
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    }
  } finally {
    creatingVersion.value = false
  }
}

onMounted(loadProjects)
</script>

<template>
  <main class="spiders-page">
    <div class="page-header">
      <h1>{{ localeStore.t('pages.spiders.title') }}</h1>
      <el-button type="primary" @click="openCreateDialog">
        {{ localeStore.t('pages.spiders.create') }}
      </el-button>
    </div>

    <div class="project-selector">
      <el-select
        v-model="selectedProjectId"
        :placeholder="localeStore.t('pages.spiders.projectId')"
        @change="loadSpiders"
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
      :data="spiders"
      stripe
      :empty-text="localeStore.t('pages.spiders.empty')"
    >
      <el-table-column
        prop="name"
        :label="localeStore.t('pages.spiders.name')"
        min-width="160"
      />
      <el-table-column
        prop="language"
        :label="localeStore.t('pages.spiders.language')"
        width="100"
      />
      <el-table-column
        prop="runtime"
        :label="localeStore.t('pages.spiders.runtime')"
        width="100"
      />
      <el-table-column
        prop="image"
        :label="localeStore.t('pages.spiders.image')"
        min-width="200"
      />
      <el-table-column :label="localeStore.t('common.actions.create')" width="120">
        <template #default="scope">
          <el-button size="small" @click="openVersionsDialog(scope.row)">
            {{ localeStore.t('pages.spiders.versions.title') }}
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

    <!-- Create Spider Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="localeStore.t('pages.spiders.create')"
    >
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="localeStore.t('pages.spiders.name')">
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.spiders.language')">
          <el-select v-model="createForm.language">
            <el-option label="Go" value="go" />
            <el-option label="Python" value="python" />
          </el-select>
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.spiders.image')">
          <el-input v-model="createForm.image" placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.spiders.command')">
          <el-input v-model="createForm.command" placeholder="./go-echo" />
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

    <!-- Versions Dialog -->
    <el-dialog
      v-model="showVersionsDialog"
      :title="localeStore.t('pages.spiders.versions.title')"
    >
      <el-table
        v-loading="loadingVersions"
        :data="versions"
        :empty-text="localeStore.t('pages.spiders.versions.empty')"
        size="small"
      >
        <el-table-column
          prop="version"
          :label="localeStore.t('pages.spiders.versions.version')"
        />
        <el-table-column
          prop="image"
          :label="localeStore.t('pages.spiders.versions.image')"
        />
      </el-table>

      <el-divider />

      <el-form :model="versionForm" inline>
        <el-form-item :label="localeStore.t('pages.spiders.versions.version')">
          <el-input v-model="versionForm.version" placeholder="1.0.0" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.spiders.versions.image')">
          <el-input v-model="versionForm.image" />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="creatingVersion"
            @click="handleCreateVersion"
          >
            {{ localeStore.t('pages.spiders.versions.create') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
  </main>
</template>

<style scoped>
.spiders-page {
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
</style>
