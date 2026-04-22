<template>
  <main class="page">
    <section class="card">
      <h1>Spiders</h1>
      <p>Register Docker-based Go or Python spiders and inspect project-scoped spider definitions.</p>
    </section>

    <section class="card">
      <h2>Create Spider</h2>
      <form class="form" @submit.prevent="submit">
        <label>
          Project ID
          <input v-model="form.projectId" required />
        </label>
        <label>
          Name
          <input v-model="form.name" required />
        </label>
        <label>
          Language
          <select v-model="form.language">
            <option value="go">go</option>
            <option value="python">python</option>
          </select>
        </label>
        <label>
          Image
          <input v-model="form.image" required placeholder="crawler/go-echo:latest" />
        </label>
        <label>
          Command
          <input v-model="form.command" placeholder="./go-echo" />
        </label>
        <button :disabled="submitting" type="submit">{{ submitting ? 'Creating...' : 'Create Spider' }}</button>
      </form>
      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="createdSpider" class="success">Created {{ createdSpider.name }} ({{ createdSpider.id }})</p>
    </section>

    <section class="card">
      <h2>Project Spiders</h2>
      <div class="toolbar">
        <input v-model="projectFilter" placeholder="project id" />
        <button :disabled="loading" @click="loadSpiders">{{ loading ? 'Loading...' : 'Load Spiders' }}</button>
      </div>
      <ul v-if="spiders.length" class="list">
        <li v-for="spider in spiders" :key="spider.id">
          <strong>{{ spider.name }}</strong>
          <span>{{ spider.language }} / {{ spider.runtime }}</span>
          <code>{{ spider.image }}</code>
        </li>
      </ul>
      <p v-else>No spiders loaded yet.</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
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
const createdSpider = ref<Spider | null>(null)
const error = ref('')
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
  error.value = ''

  try {
    createdSpider.value = await createSpider({
      projectId: form.projectId,
      name: form.name,
      language: form.language,
      runtime: 'docker',
      image: form.image,
      command: parseCommand(form.command),
    })
    await loadSpiders()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to create spider'
  } finally {
    submitting.value = false
  }
}

async function loadSpiders() {
  if (!projectFilter.value.trim()) {
    return
  }

  loading.value = true
  error.value = ''
  try {
    spiders.value = await listSpiders(projectFilter.value.trim())
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to load spiders'
  } finally {
    loading.value = false
  }
}
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

.form {
  display: grid;
  gap: 0.75rem;
  max-width: 32rem;
}

.toolbar {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.list {
  display: grid;
  gap: 0.5rem;
  padding-left: 1rem;
}

.error {
  color: #b42318;
}

.success {
  color: #027a48;
}
</style>
