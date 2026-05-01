import { defineStore } from 'pinia'

export const AUTH_TOKEN_STORAGE_KEY = 'crawler_platform_auth_token'

export function loadPersistedToken() {
  if (typeof localStorage === 'undefined' || typeof localStorage.getItem !== 'function') {
    return ''
  }
  return localStorage.getItem(AUTH_TOKEN_STORAGE_KEY) ?? ''
}

export const useAuthStore = defineStore('auth', {
  state: () => ({ token: '' }),
  actions: {
    hydrateToken() {
      this.token = loadPersistedToken()
    },
    setToken(token: string) {
      this.token = token
      if (typeof localStorage !== 'undefined' && typeof localStorage.setItem === 'function') {
        localStorage.setItem(AUTH_TOKEN_STORAGE_KEY, token)
      }
    },
    clearToken() {
      this.token = ''
      if (typeof localStorage !== 'undefined' && typeof localStorage.removeItem === 'function') {
        localStorage.removeItem(AUTH_TOKEN_STORAGE_KEY)
      }
    },
  },
})
