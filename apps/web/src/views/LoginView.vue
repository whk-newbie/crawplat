<template>
  <main class="page">
    <section class="card login-card">
      <h1>Login</h1>
      <p>Use IAM credentials to enter the crawler control plane.</p>

      <form class="form" @submit.prevent="submit">
        <label>
          Username
          <input v-model="form.username" name="username" required autocomplete="username" />
        </label>
        <label>
          Password
          <input v-model="form.password" name="password" type="password" required autocomplete="current-password" />
        </label>
        <button :disabled="submitting" type="submit">
          {{ submitting ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <p v-if="error" class="error">{{ error }}</p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  username: '',
  password: '',
})

const submitting = ref(false)
const error = ref('')

async function submit() {
  submitting.value = true
  error.value = ''

  try {
    const result = await login({
      username: form.username,
      password: form.password,
    })
    authStore.setToken(result.token)
    await router.push('/projects')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'login failed'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page {
  min-height: calc(100vh - 2rem);
  display: grid;
  place-items: center;
  padding: 1rem;
}

.card {
  border: 1px solid #d0d7de;
  border-radius: 8px;
  padding: 1rem;
  width: min(28rem, 100%);
}

.login-card {
  display: grid;
  gap: 0.75rem;
}

.form {
  display: grid;
  gap: 0.75rem;
}

.error {
  color: #b42318;
}
</style>
