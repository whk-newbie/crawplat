import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ProjectsView from '../ProjectsView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('projects view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('loads projects and creates a project via dialog', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          { id: 'p1', code: 'core-crawlers', name: 'Core Crawlers' },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({ id: 'p2', code: 'data-team', name: 'Data Team' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          { id: 'p1', code: 'core-crawlers', name: 'Core Crawlers' },
          { id: 'p2', code: 'data-team', name: 'Data Team' },
        ]),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(ProjectsView).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/projects', expect.any(Object))
    expect(container.textContent).toContain('Core Crawlers')

    const openCreateButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Create Project'))
    ;(openCreateButton as HTMLButtonElement).click()
    await nextTick()

    ;(container.querySelector('input[name="code"]') as HTMLInputElement).value = 'data-team'
    ;(container.querySelector('input[name="code"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('input[name="name"]') as HTMLInputElement).value = 'Data Team'
    ;(container.querySelector('input[name="name"]') as HTMLInputElement).dispatchEvent(new Event('input'))
    ;(container.querySelector('.dialog form') as HTMLFormElement).dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))

    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/api/v1/projects',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          code: 'data-team',
          name: 'Data Team',
        }),
      }),
    )
    expect(fetchMock).toHaveBeenNthCalledWith(3, '/api/v1/projects', expect.any(Object))
    expect(container.textContent).toContain('Data Team')
  })
})
