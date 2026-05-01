import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import NodesView from '../NodesView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('nodes view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('loads node list and renders selected node detail', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          {
            id: 'node-a',
            name: 'node-a',
            status: 'online',
            capabilities: ['docker', 'go'],
            lastSeenAt: '2026-05-01T09:30:00Z',
          },
          {
            id: 'node-b',
            name: 'node-b',
            status: 'offline',
            capabilities: ['python'],
            lastSeenAt: '2026-04-30T22:00:00Z',
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          node: {
            id: 'node-a',
            name: 'node-a',
            status: 'online',
            capabilities: ['docker', 'go'],
            lastSeenAt: '2026-05-01T09:30:00Z',
          },
          heartbeatHistory: [
            {
              seenAt: '2026-05-01T09:30:00Z',
              capabilities: ['docker', 'go'],
            },
            {
              seenAt: '2026-05-01T09:25:00Z',
              capabilities: ['docker', 'go'],
            },
          ],
          recentExecutions: [
            {
              id: 'exec-1',
              spiderId: 'spider-a',
              status: 'running',
            },
          ],
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          sessions: [
            {
              startedAt: '2026-05-01T09:00:00Z',
              endedAt: '2026-05-01T09:40:00Z',
              durationSeconds: 2400,
              heartbeatCount: 12,
            },
          ],
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(NodesView).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/nodes', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/nodes/node-a?executionLimit=20', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      '/api/v1/nodes/node-a/sessions?limit=20&gapSeconds=90',
      expect.any(Object),
    )
    expect(container.textContent).toContain('Nodes')
    expect(container.textContent).toContain('node-a')
    expect(container.textContent).toContain('node-b')
    expect(container.textContent).toContain('docker')
    expect(container.textContent).toContain('exec-1')
    expect(container.textContent).toContain('Online Sessions')
    expect(container.textContent).toContain('heartbeats: 12')
  })

  it('applies execution filter and pagination query params', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          {
            id: 'node-a',
            name: 'node-a',
            status: 'online',
            capabilities: ['docker'],
            lastSeenAt: '2026-05-01T09:30:00Z',
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          node: {
            id: 'node-a',
            name: 'node-a',
            status: 'online',
            capabilities: ['docker'],
            lastSeenAt: '2026-05-01T09:30:00Z',
          },
          heartbeatHistory: [],
          recentExecutions: [],
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ sessions: [] }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          node: {
            id: 'node-a',
            name: 'node-a',
            status: 'online',
            capabilities: ['docker'],
            lastSeenAt: '2026-05-01T09:30:00Z',
          },
          heartbeatHistory: [],
          recentExecutions: [],
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          node: {
            id: 'node-a',
            name: 'node-a',
            status: 'online',
            capabilities: ['docker'],
            lastSeenAt: '2026-05-01T09:30:00Z',
          },
          heartbeatHistory: [],
          recentExecutions: [],
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(NodesView).mount(container)
    await flushPromises()

    const statusSelect = container.querySelector('select') as HTMLSelectElement
    statusSelect.value = 'running'
    statusSelect.dispatchEvent(new Event('change'))
    await flushPromises()

    const nextButton = Array.from(container.querySelectorAll('button')).find((item) => item.textContent?.includes('Next'))
    nextButton?.dispatchEvent(new Event('click'))
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/nodes/node-a?executionLimit=20&executionStatus=running',
      expect.any(Object),
    )
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/nodes/node-a?executionLimit=20&executionOffset=20&executionStatus=running',
      expect.any(Object),
    )
  })
})
