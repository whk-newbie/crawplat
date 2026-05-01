import { createApp, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import ElementPlus from 'element-plus'
import LoginView from '../LoginView.vue'
import { useAuthStore } from '../../stores/auth'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('login view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('stores token and redirects to projects after login success', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ token: 'token-1' }),
    })

    vi.stubGlobal('fetch', fetchMock)

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/login', component: LoginView },
        { path: '/projects', component: { template: '<div>Projects page</div>' } },
      ],
    })
    const pinia = createPinia()

    await router.push('/login')
    await router.isReady()

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(LoginView).use(router).use(pinia).use(ElementPlus).mount(container)

    await flushPromises()

    const usernameInput = container.querySelector('input[name="username"]') as HTMLInputElement
    const passwordInput = container.querySelector('input[name="password"]') as HTMLInputElement

    usernameInput.value = 'admin'
    usernameInput.dispatchEvent(new Event('input'))
    passwordInput.value = 'admin123'
    passwordInput.dispatchEvent(new Event('input'))

    const submitButton = container.querySelector('.el-button--primary') as HTMLButtonElement
    submitButton.click()

    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/auth/login',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          username: 'admin',
          password: 'admin123',
        }),
      }),
    )

    const authStore = useAuthStore(pinia)
    expect(authStore.token).toBe('token-1')
    expect(router.currentRoute.value.fullPath).toBe('/projects')
  })
})
