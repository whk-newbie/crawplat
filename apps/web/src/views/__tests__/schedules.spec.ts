import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
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

  it('loads schedules and renders retry configuration', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ([
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
      ]),
    })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(SchedulesView).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/schedules', expect.any(Object))
    expect(container.textContent).toContain('nightly-go-echo')
    expect(container.textContent).toContain('*/5 * * * *')
    expect(container.textContent).toContain('retry 2')
    expect(container.textContent).toContain('30s')
  })

  it('creates a schedule and prepends it to the list', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: 'schedule-2',
          projectId: 'project-1',
          spiderId: 'spider-2',
          name: 'hourly-python',
          cronExpr: '0 * * * *',
          enabled: true,
          image: 'crawler/python-echo:latest',
          command: ['python', 'main.py'],
          retryLimit: 1,
          retryDelaySeconds: 15,
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(SchedulesView).mount(container)
    await flushPromises()

    ;(container.querySelector('input[name="projectId"]') as HTMLInputElement).value = 'project-1'
    ;(container.querySelector('input[name="projectId"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="spiderId"]') as HTMLInputElement).value = 'spider-2'
    ;(container.querySelector('input[name="spiderId"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="name"]') as HTMLInputElement).value = 'hourly-python'
    ;(container.querySelector('input[name="name"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="cronExpr"]') as HTMLInputElement).value = '0 * * * *'
    ;(container.querySelector('input[name="cronExpr"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="image"]') as HTMLInputElement).value = 'crawler/python-echo:latest'
    ;(container.querySelector('input[name="image"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="command"]') as HTMLInputElement).value = 'python main.py'
    ;(container.querySelector('input[name="command"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="retryLimit"]') as HTMLInputElement).value = '1'
    ;(container.querySelector('input[name="retryLimit"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="retryDelaySeconds"]') as HTMLInputElement).value = '15'
    ;(container.querySelector('input[name="retryDelaySeconds"]') as HTMLInputElement).dispatchEvent(new Event('input'))

    ;(container.querySelector('form') as HTMLFormElement).dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/api/v1/schedules',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          projectId: 'project-1',
          spiderId: 'spider-2',
          name: 'hourly-python',
          cronExpr: '0 * * * *',
          enabled: true,
          image: 'crawler/python-echo:latest',
          command: ['python', 'main.py'],
          retryLimit: 1,
          retryDelaySeconds: 15,
        }),
      }),
    )
    expect(container.textContent).toContain('hourly-python')
  })
})
