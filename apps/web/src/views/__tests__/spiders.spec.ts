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

  it('creates spider version with registry auth ref', async () => {
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
            id: 'v1',
            spiderId: 'spider-1',
            version: 1,
            registryAuthRef: '',
            image: 'crawler/go:latest',
            command: ['./crawler'],
            createdAt: '2026-01-01T00:00:00Z',
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => (['ghcr-prod']),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: 'v2',
          spiderId: 'spider-1',
          version: 2,
          registryAuthRef: 'ghcr-prod',
          image: 'crawler/go:v2',
          command: ['./crawler', '--fast'],
          createdAt: '2026-01-02T00:00:00Z',
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
            registryAuthRef: 'ghcr-prod',
            image: 'crawler/go:v2',
            command: ['./crawler', '--fast'],
            createdAt: '2026-01-02T00:00:00Z',
          },
          {
            id: 'v1',
            spiderId: 'spider-1',
            version: 1,
            registryAuthRef: '',
            image: 'crawler/go:latest',
            command: ['./crawler'],
            createdAt: '2026-01-01T00:00:00Z',
          },
        ]),
      })
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
            image: 'crawler/go:v2',
            command: ['./crawler', '--fast'],
          }],
          total: 1,
          limit: 20,
          offset: 0,
        }),
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

    const loadRefsButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Load Refs'))
    ;(loadRefsButton as HTMLButtonElement).click()
    await flushPromises()

    const inputs = [...container.querySelectorAll('input')]
    const imageInput = inputs.find((input) => input.getAttribute('placeholder') === 'crawler/go:v2')
    const commandInput = inputs.find((input) => input.getAttribute('placeholder') === './crawler --fast')
    ;(imageInput as HTMLInputElement).value = 'crawler/go:v2'
    imageInput?.dispatchEvent(new Event('input'))
    ;(commandInput as HTMLInputElement).value = './crawler --fast'
    commandInput?.dispatchEvent(new Event('input'))
    await flushPromises()

    const createVersionButton = [...container.querySelectorAll('button')].find((button) =>
      button.textContent?.includes('Create Version')
    )
    ;(createVersionButton as HTMLButtonElement).click()
    await flushPromises()

    const createCall = fetchMock.mock.calls[3]
    expect(createCall[0]).toBe('/api/v1/spiders/spider-1/versions')
    expect(createCall[1]).toMatchObject({
      method: 'POST',
    })
    expect(JSON.parse(createCall[1].body as string)).toMatchObject({
      registryAuthRef: 'ghcr-prod',
      image: 'crawler/go:v2',
      command: ['./crawler', '--fast'],
    })
    expect(container.textContent).toContain('ghcr-prod')
  })
})
