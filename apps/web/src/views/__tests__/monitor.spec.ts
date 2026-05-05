import { createApp, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
import MonitorView from '../MonitorView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('monitor view', () => {
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

  it('loads and renders monitor overview counters', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        executions: { total: 12, pending: 7, running: 3, failed: 1, succeeded: 1 },
        nodes: { total: 4, online: 2, offline: 2 },
        generatedAt: '2026-04-23T08:00:00Z',
      }),
    })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(MonitorView).use(createPinia()).use(ElementPlus).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/monitor/overview', expect.any(Object))
    expect(container.textContent).toContain('总执行数')
    expect(container.textContent).toContain('12')
    expect(container.textContent).toContain('等待中执行')
    expect(container.textContent).toContain('7')
    expect(container.textContent).toContain('在线节点')
    expect(container.textContent).toContain('2')
  })

  it('shows an error when the overview request fails', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 503,
      text: async () => 'monitor unavailable',
    })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(MonitorView).use(createPinia()).use(ElementPlus).mount(container)
    await flushPromises()

    expect(container.textContent).toContain('request failed: 503')
  })
})
