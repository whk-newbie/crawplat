import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
import DatasourcesView from '../DatasourcesView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

const findButtonByText = (root: HTMLElement | Document, text: string): HTMLButtonElement | null =>
  [...root.querySelectorAll('button')].find((button) =>
    button.textContent?.toLowerCase().includes(text.toLowerCase()),
  ) as HTMLButtonElement | null

describe('datasources view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('loads datasource list on initial render', async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ([
        {
          id: 'ds-1',
          projectId: 'project-1',
          name: 'main-pg',
          type: 'postgresql',
          readonly: true,
          config: { schema: 'public' },
        },
      ]),
    })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(DatasourcesView).use(ElementPlus).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/datasources', expect.any(Object))
    expect(container.textContent).toContain('main-pg')
    expect(container.textContent).toContain('postgresql')
  })

  it('runs test and preview then renders result content', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          {
            id: 'ds-1',
            projectId: 'project-1',
            name: 'main-pg',
            type: 'postgresql',
            readonly: true,
            config: { schema: 'public' },
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          datasourceId: 'ds-1',
          status: 'ok',
          message: 'mock connection test passed',
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          datasourceId: 'ds-1',
          datasourceType: 'postgresql',
          rows: [{ id: 'sample-1', name: 'example' }],
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(DatasourcesView).use(ElementPlus).mount(container)
    await flushPromises()

    const testButton = findButtonByText(container, 'test')
    expect(testButton).toBeTruthy()
    ;(testButton as HTMLButtonElement).click()
    await flushPromises()
    expect(container.textContent).toContain('mock connection test passed')

    const previewButton = findButtonByText(container, 'preview')
    expect(previewButton).toBeTruthy()
    ;(previewButton as HTMLButtonElement).click()
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/datasources/ds-1/test', expect.objectContaining({ method: 'POST' }))
    expect(fetchMock).toHaveBeenNthCalledWith(3, '/api/v1/datasources/ds-1/preview', expect.objectContaining({ method: 'POST' }))
    expect(container.textContent).toContain('sample-1')
    expect(container.textContent).toContain('example')
  })
})
