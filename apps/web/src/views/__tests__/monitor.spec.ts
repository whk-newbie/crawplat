import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
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
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('loads overview, rules and events', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          executions: { total: 12, pending: 7, running: 3, failed: 1, succeeded: 1 },
          nodes: { total: 4, online: 2, offline: 2 },
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([{
          id: 'r1',
          name: 'node offline',
          ruleType: 'node_offline',
          enabled: true,
          webhookUrl: 'https://example.com',
          cooldownSeconds: 60,
          timeoutSeconds: 3,
          offlineGraceSeconds: 60,
          createdAt: '2026-05-01T00:00:00Z',
          updatedAt: '2026-05-01T00:00:00Z',
        }]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          items: [{
            id: 'e1',
            ruleType: 'node_offline',
            entityId: 'mvp-node',
            deliveryStatus: 'failed',
            createdAt: '2026-05-01T00:00:00Z',
          }],
          total: 1,
          limit: 20,
          offset: 0,
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(MonitorView).use(ElementPlus).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/monitor/overview', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/monitor/alerts/rules', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(3, '/api/v1/monitor/alerts/events?limit=20&offset=0', expect.any(Object))

    expect(container.textContent).toContain('Alert Rules')
    expect(container.textContent).toContain('node offline')
    expect(container.textContent).toContain('Recent Alert Events')
    expect(container.textContent).toContain('mvp-node')
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

    createApp(MonitorView).use(ElementPlus).mount(container)
    await flushPromises()

    expect(container.textContent).toContain('monitor unavailable')
  })
})
