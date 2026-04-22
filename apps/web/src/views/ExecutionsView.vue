<template>
  <main class="page">
    <section class="card">
      <h1>Executions</h1>
      <p>Create a manual execution and jump into its detail page, or inspect an existing execution by ID.</p>
    </section>

    <section class="card">
      <h2>Create Execution</h2>
      <form class="form" @submit.prevent="submit">
        <label>
          Project ID
          <input v-model="form.projectId" required />
        </label>
        <label>
          Spider ID
          <input v-model="form.spiderId" required />
        </label>
        <label>
          Image
          <input v-model="form.image" required placeholder="crawler/go-echo:latest" />
        </label>
        <label>
          Command
          <input v-model="form.command" placeholder="./go-echo" />
        </label>
        <button :disabled="submitting" type="submit">{{ submitting ? 'Creating...' : 'Create Execution' }}</button>
      </form>
      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="createdExecutionId" class="success">
        Created execution
        <router-link :to="`/executions/${createdExecutionId}`">{{ createdExecutionId }}</router-link>
      </p>
    </section>

    <section class="card">
      <h2>Open Execution Detail</h2>
      <div class="toolbar">
        <input v-model="lookupId" placeholder="execution id" />
        <button @click="openExecution">Open</button>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createExecution } from '../api/executions'

const router = useRouter()

const form = reactive({
  projectId: 'project-1',
  spiderId: '',
  image: '',
  command: '',
})

const lookupId = ref('')
const createdExecutionId = ref('')
const error = ref('')
const submitting = ref(false)

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
    const execution = await createExecution({
      projectId: form.projectId,
      spiderId: form.spiderId,
      image: form.image,
      command: parseCommand(form.command),
    })
    createdExecutionId.value = execution.id
    lookupId.value = execution.id
    await router.push(`/executions/${execution.id}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to create execution'
  } finally {
    submitting.value = false
  }
}

async function openExecution() {
  if (!lookupId.value.trim()) {
    return
  }
  await router.push(`/executions/${lookupId.value.trim()}`)
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
}

.error {
  color: #b42318;
}

.success {
  color: #027a48;
}
</style>
