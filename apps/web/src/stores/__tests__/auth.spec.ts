import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '../auth'

describe('auth store', () => {
  let storage: Record<string, string> = {}

  beforeEach(() => {
    setActivePinia(createPinia())
    storage = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => storage[key] ?? null,
      setItem: (key: string, value: string) => {
        storage[key] = value
      },
      removeItem: (key: string) => {
        delete storage[key]
      },
      clear: () => {
        storage = {}
      },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('stores token after login success', async () => {
    const store = useAuthStore()
    store.setToken('token-1')
    expect(store.token).toBe('token-1')
    expect(localStorage.getItem('crawler_platform_auth_token')).toBe('token-1')
  })

  it('hydrates token from localStorage', async () => {
    localStorage.setItem('crawler_platform_auth_token', 'persisted-token')
    const store = useAuthStore()
    store.hydrateToken()
    expect(store.token).toBe('persisted-token')
  })

  it('clears token from state and localStorage', async () => {
    const store = useAuthStore()
    store.setToken('token-1')
    store.clearToken()
    expect(store.token).toBe('')
    expect(localStorage.getItem('crawler_platform_auth_token')).toBeNull()
  })
})
