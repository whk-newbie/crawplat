import { defineStore } from 'pinia'

const tokenKey = 'crawler_platform_token'

function safeStore(): Storage | null {
  try {
    if (typeof localStorage !== 'undefined' && localStorage.getItem) {
      return localStorage
    }
  } catch { /* not available */ }
  return null
}

function readToken(): string {
  return safeStore()?.getItem(tokenKey) ?? ''
}

function persistToken(token: string) {
  const store = safeStore()
  if (!store) return
  if (token) {
    store.setItem(tokenKey, token)
  } else {
    store.removeItem(tokenKey)
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({ token: readToken() }),
  actions: {
    setToken(token: string) {
      this.token = token
      persistToken(token)
    },
    clearToken() {
      this.token = ''
      persistToken('')
    },
  },
})
