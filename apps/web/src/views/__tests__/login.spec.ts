import { createApp, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
import LoginView from '../LoginView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('login view', () => {
  let storage: Record<string, string>

  beforeEach(() => {
    storage = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => storage[key] ?? null,
      setItem: (key: string, value: string) => { storage[key] = value },
      removeItem: (key: string) => { delete storage[key] },
      clear: () => { storage = {} },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('renders the login page title via i18n', async () => {
    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(LoginView).use(createPinia()).use(ElementPlus).mount(container)
    await flushPromises()

    expect(container.textContent).toContain('登录')
    expect(container.textContent).toContain('Web MVP')
  })

  it('renders English title when locale is en-US', async () => {
    storage['crawler_platform_locale'] = 'en-US'

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(LoginView).use(createPinia()).use(ElementPlus).mount(container)
    await flushPromises()

    expect(container.textContent).toContain('Login')
    expect(container.textContent).toContain('Web MVP')
  })
})
