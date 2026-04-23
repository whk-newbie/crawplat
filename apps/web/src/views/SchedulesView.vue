<template>
  <main class="page">
    <section class="card">
      <h1>Schedules</h1>
      <p>Create cron-based spider runs and inspect the current retry policy attached to each schedule.</p>
    </section>

    <section class="card">
      <h2>Create Schedule</h2>
      <form class="form" @submit.prevent="submit">
        <label>
          Project ID
          <input v-model="form.projectId" name="projectId" required />
        </label>
        <label>
          Spider ID
          <input v-model="form.spiderId" name="spiderId" required />
        </label>
        <label>
          Name
          <input v-model="form.name" name="name" required />
        </label>
        <label>
          Cron Expression
          <input v-model="form.cronExpr" name="cronExpr" required placeholder="*/5 * * * *" />
        </label>
        <label>
          Image
          <input v-model="form.image" name="image" required placeholder="crawler/go-echo:latest" />
        </label>
        <label>
          Command
          <input v-model="form.command" name="command" placeholder="./go-echo" />
        </label>
        <label>
          Retry Limit
          <input v-model.number="form.retryLimit" name="retryLimit" min="0" type="number" />
        </label>
        <label>
          Retry Delay Seconds
          <input v-model.number="form.retryDelaySeconds" name="retryDelaySeconds" min="0" type="number" />
        </label>
        <label class="checkbox">
          <input v-model="form.enabled" type="checkbox" />
          Enabled
        </label>
        <button :disabled="submitting" type="submit">
          {{ submitting ? 'Creating...' : 'Create Schedule' }}
        </button>
      </form>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <section class="card">
      <div class="toolbar">
        <h2>Current Schedules</h2>
        <button :disabled="loading" @click="loadSchedules">
          {{ loading ? 'Refreshing...' : 'Refresh' }}
        </button>
      </div>
      <p v-if="loading">Loading schedules...</p>
      <p v-else-if="schedules.length === 0">No schedules yet.</p>
      <ul v-else class="schedule-list">
        <li v-for="schedule in schedules" :key="schedule.id">
          <div class="row">
            <strong>{{ schedule.name }}</strong>
            <span :class="schedule.enabled ? 'enabled' : 'disabled'">
              {{ schedule.enabled ? 'enabled' : 'disabled' }}
            </span>
          </div>
          <div class="meta">
            <span>{{ schedule.cronExpr }}</span>
            <span>{{ schedule.image }}</span>
            <span>retry {{ schedule.retryLimit }}</span>
            <span>{{ schedule.retryDelaySeconds }}s</span>
          </div>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { createSchedule, listSchedules, type Schedule } from '../api/schedules'

const schedules = ref<Schedule[]>([])
const loading = ref(true)
const submitting = ref(false)
const error = ref('')

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
  error.value = ''
  try {
    schedules.value = await listSchedules()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to load schedules'
  } finally {
    loading.value = false
  }
}

async function submit() {
  submitting.value = true
  error.value = ''
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
    schedules.value = [schedule, ...schedules.value]
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'failed to create schedule'
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  void loadSchedules()
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

.form {
  display: grid;
  gap: 0.75rem;
  max-width: 32rem;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.schedule-list {
  display: grid;
  gap: 0.75rem;
  padding: 0;
  list-style: none;
}

.schedule-list li {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 0.75rem;
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.5rem;
  color: #57606a;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.enabled {
  color: #027a48;
}

.disabled {
  color: #b42318;
}

.error {
  color: #b42318;
}
</style>
