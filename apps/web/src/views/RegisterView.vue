<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register } from '../api/auth'
import { ApiError } from '../api/client'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
const router = useRouter()

const form = reactive({ username: '', password: '' })
const submitting = ref(false)

async function handleRegister() {
  if (!form.username.trim()) {
    ElMessage.warning(localeStore.t('pages.register.errors.usernameRequired'))
    return
  }
  if (!form.password) {
    ElMessage.warning(localeStore.t('pages.register.errors.passwordRequired'))
    return
  }

  submitting.value = true
  try {
    await register({ username: form.username.trim(), password: form.password })
    ElMessage.success(localeStore.t('pages.register.success'))
    router.push('/login')
  } catch (err) {
    if (err instanceof ApiError) {
      ElMessage.error(localeStore.t(err.code))
    } else {
      ElMessage.error(localeStore.t('errors.unknown'))
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="auth-page">
    <el-card class="auth-card">
      <template #header>
        <h1>{{ localeStore.t('pages.register.title') }}</h1>
      </template>

      <el-form
        :model="form"
        label-position="top"
        @keyup.enter="handleRegister"
      >
        <el-form-item :label="localeStore.t('pages.register.username')">
          <el-input
            v-model="form.username"
            autocomplete="username"
          />
        </el-form-item>

        <el-form-item :label="localeStore.t('pages.register.password')">
          <el-input
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            show-password
          />
        </el-form-item>

        <el-button
          type="primary"
          :loading="submitting"
          class="submit-btn"
          @click="handleRegister"
        >
          {{ localeStore.t('pages.register.submit') }}
        </el-button>
      </el-form>

      <p class="auth-switch">
        <router-link to="/login">
          {{ localeStore.t('pages.register.toLogin') }}
        </router-link>
      </p>
    </el-card>
  </main>
</template>

<style scoped>
.auth-page {
  align-items: center;
  display: flex;
  justify-content: center;
  min-height: 60vh;
  padding: 2rem;
}

.auth-card {
  max-width: 400px;
  width: 100%;
}

.submit-btn {
  width: 100%;
}

.auth-switch {
  margin-top: 1rem;
  text-align: center;
}
</style>
