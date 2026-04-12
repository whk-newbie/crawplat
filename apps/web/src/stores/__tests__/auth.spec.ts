import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '../auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('stores token after login success', async () => {
    const store = useAuthStore()
    store.setToken('token-1')
    expect(store.token).toBe('token-1')
  })
})
