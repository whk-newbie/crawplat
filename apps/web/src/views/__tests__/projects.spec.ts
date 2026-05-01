import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ElementPlus from 'element-plus'
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

  it('loads projects on mount and displays them', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          { id: 'p1', code: 'core-crawlers', name: 'Core Crawlers' },
        ]),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(ProjectsView).use(ElementPlus).mount(container)
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/projects', expect.any(Object))
    expect(container.textContent).toContain('Core Crawlers')
  })

  it('creates a project via dialog', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([]),
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
          { id: 'p2', code: 'data-team', name: 'Data Team' },
        ]),
      })

    vi.stubGlobal('fetch', fetchMock)

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(ProjectsView).use(ElementPlus).mount(container)
    await flushPromises()

    const buttons = [...container.querySelectorAll('button')]
    const createButton = buttons.find((b) => b.textContent?.includes('Create Project'))
    ;(createButton as HTMLButtonElement).click()
    await flushPromises()

    const codeInput = document.querySelector('input[name="code"]') as HTMLInputElement
    const nameInput = document.querySelector('input[name="name"]') as HTMLInputElement

    codeInput.value = 'data-team'
    codeInput.dispatchEvent(new Event('input'))
    nameInput.value = 'Data Team'
    nameInput.dispatchEvent(new Event('input'))

    const dialogButtons = [...document.querySelectorAll('.el-dialog__footer button')]
    const confirmButton = dialogButtons.find((b) => b.textContent?.includes('Create'))
    ;(confirmButton as HTMLButtonElement).click()

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
  })
})
