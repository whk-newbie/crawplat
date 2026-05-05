import { createApp, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import ElementPlus from 'element-plus'
import ExecutionDetailView from '../ExecutionDetailView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('execution detail view', () => {
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

  it('renders execution details after fetching', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          id: 'exec-1',
          projectId: 'project-1',
          spiderId: 'spider-1',
          status: 'running',
          triggerSource: 'manual',
          image: 'crawler/go-echo:latest',
          command: ['./go-echo'],
          createdAt: '2026-04-22T15:00:00Z',
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => [{
          id: 'log-1',
          executionId: 'exec-1',
          message: 'started',
          createdAt: '2026-04-22T15:00:01Z',
        }],
      })

    vi.stubGlobal('fetch', fetchMock)

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/executions/:id', component: ExecutionDetailView }],
    })

    await router.push('/executions/exec-1')
    await router.isReady()

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(ExecutionDetailView).use(router).use(createPinia()).use(ElementPlus).mount(container)
    await flushPromises()

    expect(container.textContent).toContain('exec-1')
    expect(container.textContent).toContain('running')
    expect(container.textContent).toContain('started')
  })
})
