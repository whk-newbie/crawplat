<template>
  <div class="page">
    <el-card>
      <template #header>
        <h1>{{ localeStore.t('pages.executions.title') }}</h1>
      </template>
      <p>{{ localeStore.t('pages.executions.placeholder') }}</p>
    </el-card>

    <el-card>
      <template #header>
        <h2>{{ localeStore.t('pages.executions.createTitle') }}</h2>
      </template>
      <el-form label-width="120px" @submit.prevent="submit">
        <el-form-item :label="localeStore.t('pages.executions.projectId')">
          <el-input v-model="form.projectId" required />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.spiderId')">
          <el-input v-model="form.spiderId" required />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.image')">
          <el-input v-model="form.image" required placeholder="crawler/go-echo:latest" />
        </el-form-item>
        <el-form-item :label="localeStore.t('pages.executions.command')">
          <el-input v-model="form.command" placeholder="./go-echo" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" native-type="submit">
            {{ submitting ? localeStore.t('pages.executions.creating') : localeStore.t('pages.executions.createAction') }}
          </el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="error" :title="error" type="error" show-icon closable @close="error = ''" />
      <el-alert v-if="createdExecutionId" type="success" show-icon>
        <template #title>
          {{ localeStore.t('pages.executions.createdMessage') }}
          <router-link :to="`/executions/${createdExecutionId}`">{{ createdExecutionId }}</router-link>
        </template>
      </el-alert>
    </el-card>

    <el-card>
      <template #header>
        <h2>{{ localeStore.t('pages.executions.lookupTitle') }}</h2>
      </template>
      <div class="toolbar">
        <el-input
          v-model="lookupId"
          :placeholder="localeStore.t('pages.executions.lookupPlaceholder')"
          style="max-width: 300px"
        />
        <el-button @click="openExecution">{{ localeStore.t('pages.executions.openAction') }}</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLocaleStore } from '../stores/locale'
import { createExecution } from '../api/executions'

const router = useRouter()
const localeStore = useLocaleStore()

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
}
.toolbar {
  align-items: center;
  display: flex;
  gap: 0.75rem;
}
</style>
