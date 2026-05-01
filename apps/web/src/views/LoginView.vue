<template>
  <div class="login-page">
    <el-card class="login-card">
      <template #header>
        <h2>Crawler Platform</h2>
        <p class="login-subtitle">Use IAM credentials to enter the control plane.</p>
      </template>
      <el-form :model="form" @submit.prevent="submit">
        <el-form-item label="Username">
          <el-input
            v-model="form.username"
            name="username"
            autocomplete="username"
            placeholder="Enter username"
          />
        </el-form-item>
        <el-form-item label="Password">
          <el-input
            v-model="form.password"
            name="password"
            type="password"
            show-password
            autocomplete="current-password"
            placeholder="Enter password"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="submitting" style="width: 100%">
            Sign In
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  username: '',
  password: '',
})

const submitting = ref(false)

async function submit() {
  submitting.value = true

  try {
    const result = await login({
      username: form.username,
      password: form.password,
    })
    authStore.setToken(result.token)
    ElMessage.success('Login successful')
    await router.push('/projects')
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : 'login failed')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1rem;
  background: var(--el-bg-color-page);
}

.login-card {
  width: min(400px, 100%);
}

.login-subtitle {
  margin: 4px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
</style>
