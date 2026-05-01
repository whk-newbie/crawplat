<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
      <div>
        <h2 style="margin: 0">Executions</h2>
        <p style="margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 14px">
          Create a manual execution and jump into its detail page, or inspect an existing execution by ID.
        </p>
      </div>
      <el-button type="primary" @click="createDialogVisible = true">Create Execution</el-button>
    </div>

    <el-card>
      <template #header>Open Execution Detail</template>
      <div style="display: flex; gap: 12px">
        <el-input v-model="lookupId" placeholder="execution id" style="width: 360px" />
        <el-button @click="openExecution">Open</el-button>
      </div>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="Create Execution" width="500px">
      <el-form :model="form" label-position="top" @submit.prevent="submit">
        <el-form-item label="Project ID" required>
          <el-input v-model="form.projectId" />
        </el-form-item>
        <el-form-item label="Spider ID" required>
          <el-input v-model="form.spiderId" />
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
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createExecution } from '../api/executions'

const router = useRouter()

const form = reactive({
  projectId: 'project-1',
  spiderId: '',
  image: '',
  command: '',
})

const lookupId = ref('')
const createDialogVisible = ref(false)
const submitting = ref(false)

function parseCommand(input: string) {
  return input
    .split(' ')
    .map((item) => item.trim())
    .filter(Boolean)
}

async function submit() {
  submitting.value = true
  try {
    const execution = await createExecution({
      projectId: form.projectId,
      spiderId: form.spiderId,
      image: form.image,
      command: parseCommand(form.command),
    })
    ElMessage.success('Execution created')
    createDialogVisible.value = false
    lookupId.value = execution.id
    await router.push(`/executions/${execution.id}`)
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'failed to create execution')
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
