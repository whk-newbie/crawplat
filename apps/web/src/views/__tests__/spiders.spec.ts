import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
import SpidersView from '../SpidersView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('spiders view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('opens versions dialog and requests spider versions', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          items: [{
            id: 'spider-1',
            projectId: 'project-1',
            name: 'crawler-a',
            language: 'go',
            runtime: 'docker',
            image: 'crawler/go:latest',
            command: ['./crawler'],
          }],
          total: 1,
          limit: 20,
          offset: 0,
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          {
            id: 'v2',
            spiderId: 'spider-1',
            version: 2,
            image: 'crawler/go:v2',
            command: ['./crawler', '--fast'],
            createdAt: '2026-01-02T00:00:00Z',
          },
          {
            id: 'v1',
            spiderId: 'spider-1',
            version: 1,
            image: 'crawler/go:latest',
            command: ['./crawler'],
            createdAt: '2026-01-01T00:00:00Z',
          },
        ]),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(SpidersView).use(ElementPlus).mount(container)
    await flushPromises()

    const loadButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Load Spiders'))
    ;(loadButton as HTMLButtonElement).click()
    await flushPromises()

    const versionsButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Versions'))
    ;(versionsButton as HTMLButtonElement).click()
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/spiders/spider-1/versions', expect.any(Object))
    expect(container.textContent).toContain('crawler/go:v2')
  })
})
