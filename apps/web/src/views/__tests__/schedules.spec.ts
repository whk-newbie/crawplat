import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
import SchedulesView from '../SchedulesView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('schedules view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('loads schedules and renders schedule data', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        items: [
          {
            id: 'schedule-1',
            projectId: 'project-1',
            spiderId: 'spider-1',
            name: 'nightly-go-echo',
            cronExpr: '*/5 * * * *',
            enabled: true,
            image: 'crawler/go-echo:latest',
            command: ['./go-echo'],
            retryLimit: 2,
            retryDelaySeconds: 30,
          },
        ],
        total: 1,
        limit: 20,
        offset: 0,
      }),
    })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(SchedulesView).use(ElementPlus).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/schedules', expect.any(Object))
    expect(container.textContent).toContain('nightly-go-echo')
    expect(container.textContent).toContain('*/5 * * * *')
    expect(container.textContent).toContain('enabled')
  })

  it('creates schedule from loaded spider versions', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          items: [],
          total: 0,
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
            command: ['./crawler', '--v2'],
            createdAt: '2026-05-01T00:00:00Z',
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: 'schedule-2',
          projectId: 'project-1',
          spiderId: 'spider-1',
          spiderVersion: 2,
          name: 'nightly-v2',
          cronExpr: '*/10 * * * *',
          enabled: true,
          image: 'crawler/go:v2',
          command: ['./crawler', '--v2'],
          retryLimit: 1,
          retryDelaySeconds: 20,
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(SchedulesView).use(ElementPlus).mount(container)
    await flushPromises()

    const createButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Create Schedule'))
    ;(createButton as HTMLButtonElement).click()
    await flushPromises()

    const spiderInput = document.querySelector('input[name="spiderId"]') as HTMLInputElement
    const nameInput = document.querySelector('input[name="name"]') as HTMLInputElement
    const cronInput = document.querySelector('input[name="cronExpr"]') as HTMLInputElement
    spiderInput.value = 'spider-1'
    spiderInput.dispatchEvent(new Event('input'))
    nameInput.value = 'nightly-v2'
    nameInput.dispatchEvent(new Event('input'))
    cronInput.value = '*/10 * * * *'
    cronInput.dispatchEvent(new Event('input'))

    const loadVersionsButton = [...document.querySelectorAll('button')].find((button) => button.textContent?.includes('Load Spider Versions'))
    ;(loadVersionsButton as HTMLButtonElement).click()
    await flushPromises()

    const footerButtons = [...document.querySelectorAll('.el-dialog__footer button')]
    const confirmButton = footerButtons.find((button) => button.textContent?.includes('Create'))
    ;(confirmButton as HTMLButtonElement).click()
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/spiders/spider-1/versions', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      '/api/v1/schedules',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          projectId: 'project-1',
          spiderId: 'spider-1',
          spiderVersion: 2,
          name: 'nightly-v2',
          cronExpr: '*/10 * * * *',
          enabled: true,
          image: 'crawler/go:v2',
          command: ['./crawler', '--v2'],
          retryLimit: 0,
          retryDelaySeconds: 0,
        }),
      }),
    )
  })
})
