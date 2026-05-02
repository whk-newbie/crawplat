import { createApp, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import ElementPlus from 'element-plus'
import ExecutionDetailView from '../ExecutionDetailView.vue'
import ExecutionsView from '../ExecutionsView.vue'

const flushPromises = async () => {
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
  await Promise.resolve()
  await new Promise((resolve) => setTimeout(resolve, 0))
  await nextTick()
}

describe('execution detail view', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.body.innerHTML = ''
  })

  it('renders execution details after selecting an execution', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          id: 'exec-1',
          projectId: 'project-1',
          spiderId: 'spider-1',
          status: 'running',
          triggerSource: 'manual',
          image: 'crawler/go-echo:latest',
          command: ['./go-echo'],
          createdAt: '2026-04-22T15:00:00Z',
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          {
            id: 'log-1',
            executionId: 'exec-1',
            message: 'started',
            createdAt: '2026-04-22T15:00:01Z',
          },
        ]),
      })

    vi.stubGlobal('fetch', fetchMock)

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/executions/:id', component: ExecutionDetailView }],
    })

    await router.push('/executions/exec-1')
    await router.isReady()

    const container = document.createElement('div')
    document.body.appendChild(container)

    createApp(ExecutionDetailView).use(router).use(ElementPlus).mount(container)
    await flushPromises()

    expect(container.textContent).toContain('Status')
    expect(container.textContent).toContain('running')
    expect(container.textContent).toContain('started')
  })

  it('creates execution from selected spider version with derived registry auth ref', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ([
          {
            id: 'v3',
            spiderId: 'spider-1',
            version: 3,
            registryAuthRef: 'ghcr-prod',
            image: 'crawler/go:v3',
            command: ['./crawler', '--v3'],
            createdAt: '2026-05-02T00:00:00Z',
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: 'exec-2',
          projectId: 'project-1',
          spiderId: 'spider-1',
          spiderVersion: 3,
          registryAuthRef: 'ghcr-prod',
          image: 'crawler/go:v3',
          command: ['./crawler', '--v3'],
          status: 'pending',
        }),
      })

    vi.stubGlobal('fetch', fetchMock)

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: ExecutionsView },
        { path: '/executions/:id', component: ExecutionDetailView },
      ],
    })
    await router.push('/')
    await router.isReady()

    const container = document.createElement('div')
    document.body.appendChild(container)
    createApp(ExecutionsView).use(router).use(ElementPlus).mount(container)
    await flushPromises()

    const createButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Create Execution'))
    ;(createButton as HTMLButtonElement).click()
    await flushPromises()

    const inputs = [...document.querySelectorAll('input')]
    const noPlaceholderInputs = inputs.filter((input) => input.getAttribute('placeholder') == null)
    const spiderIDInput = noPlaceholderInputs[1]
    const imageInput = inputs.find((input) => input.getAttribute('placeholder') === 'crawler/go-echo:latest')
    const commandInput = inputs.find((input) => input.getAttribute('placeholder') === './go-echo')
    ;(spiderIDInput as HTMLInputElement).value = 'spider-1'
    spiderIDInput?.dispatchEvent(new Event('input'))
    await flushPromises()

    const loadVersionsButton = [...document.querySelectorAll('button')].find((button) => button.textContent?.includes('Load Spider Versions'))
    ;(loadVersionsButton as HTMLButtonElement).click()
    await flushPromises()

    expect((imageInput as HTMLInputElement).value).toBe('crawler/go:v3')
    expect((commandInput as HTMLInputElement).value).toBe('./crawler --v3')

    const confirmButton = [...document.querySelectorAll('.el-dialog__footer button')].find((button) =>
      button.textContent?.includes('Create')
    )
    ;(confirmButton as HTMLButtonElement).click()
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/spiders/spider-1/versions', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/api/v1/executions',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          projectId: 'project-1',
          spiderId: 'spider-1',
          spiderVersion: 3,
          registryAuthRef: 'ghcr-prod',
          image: 'crawler/go:v3',
          command: ['./crawler', '--v3'],
        }),
      }),
    )
  })
})
