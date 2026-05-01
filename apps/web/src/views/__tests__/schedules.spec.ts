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
})
