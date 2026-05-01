<template>
  <main class="page">
    <section class="card hero">
      <div>
        <h1>Projects</h1>
        <p>Manage project namespaces for spiders, schedules, and executions.</p>
      </div>
      <div class="toolbar">
        <button :disabled="loading" @click="loadProjects">
          {{ loading ? 'Refreshing...' : 'Refresh' }}
        </button>
        <button @click="openCreateDialog">Create Project</button>
      </div>
    </section>

    <section class="card">
      <p v-if="loading">Loading projects...</p>
      <p v-else-if="error" class="error">{{ error }}</p>
      <ul v-else-if="projects.length" class="project-list">
        <li v-for="project in projects" :key="project.id">
          <strong>{{ project.name }}</strong>
          <span>{{ project.code }}</span>
          <code>{{ project.id }}</code>
        </li>
      </ul>
      <p v-else>No projects yet.</p>
    </section>

    <section v-if="createDialogVisible" class="dialog-backdrop">
      <article class="dialog">
        <h2>Create Project</h2>
        <form class="form" @submit.prevent="submitCreateProject">
          <label>
            Code
            <input v-model="createForm.code" name="code" required placeholder="core-crawlers" />
          </label>
          <label>
            Name
            <input v-model="createForm.name" name="name" required placeholder="Core Crawlers" />
          </label>
          <div class="actions">
            <button :disabled="creating" type="button" @click="closeCreateDialog">Cancel</button>
            <button :disabled="creating" type="submit">
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
        <p v-if="createError" class="error">{{ createError }}</p>
      </article>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { createProject, listProjects, type Project } from '../api/projects'

const projects = ref<Project[]>([])
const loading = ref(true)
const creating = ref(false)
const error = ref('')
const createError = ref('')
const createDialogVisible = ref(false)

const createForm = reactive({
  code: '',
  name: '',
})

function resetCreateForm() {
  createForm.code = ''
  createForm.name = ''
  createError.value = ''
}

function openCreateDialog() {
  resetCreateForm()
  createDialogVisible.value = true
}

function closeCreateDialog() {
  createDialogVisible.value = false
  createError.value = ''
}

async function loadProjects() {
  loading.value = true
  error.value = ''
  try {
    projects.value = await listProjects()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to load projects'
  } finally {
    loading.value = false
  }
}

async function submitCreateProject() {
  creating.value = true
  createError.value = ''
  try {
    await createProject({
      code: createForm.code,
      name: createForm.name,
    })
    await loadProjects()
    closeCreateDialog()
  } catch (err) {
    createError.value = err instanceof Error ? err.message : 'failed to create project'
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  void loadProjects()
})
</script>

<style scoped>
.page {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.card {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 1rem;
}

.hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.toolbar {
  display: flex;
  gap: 0.5rem;
}

.project-list {
  display: grid;
  gap: 0.5rem;
  list-style: none;
  padding: 0;
}

.project-list li {
  display: grid;
  gap: 0.25rem;
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 0.75rem;
}

.dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgb(12 17 29 / 45%);
  display: grid;
  place-items: center;
  padding: 1rem;
}

.dialog {
  width: min(32rem, 100%);
  border: 1px solid #d0d7de;
  border-radius: 8px;
  background: #fff;
  padding: 1rem;
  display: grid;
  gap: 0.75rem;
}

.form {
  display: grid;
  gap: 0.75rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.error {
  color: #b42318;
}
</style>
