import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import router from '../index'
import { useAuthStore } from '../../stores/auth'

describe('router auth guard', () => {
  let storage: Record<string, string> = {}

  beforeEach(async () => {
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
    const store = useAuthStore()
    store.clearToken()
    await router.push('/login')
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('allows public login route without token', async () => {
    await router.push('/login')
    expect(router.currentRoute.value.fullPath).toBe('/login')
  })

  it('redirects protected routes to /login when token is missing', async () => {
    await router.push('/projects')
    expect(router.currentRoute.value.fullPath).toBe('/login')
  })

  it('allows protected routes when token exists in localStorage', async () => {
    localStorage.setItem('crawler_platform_auth_token', 'persisted-token')
    await router.push('/projects')
    expect(router.currentRoute.value.fullPath).toBe('/projects')
  })
})
