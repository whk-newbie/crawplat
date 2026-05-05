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
        json: async () => ({
          items: [{ id: 'project-1', code: 'project-1', name: 'Project One' }],
          total: 1,
          limit: 20,
          offset: 0,
        }),
      })
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

    const dialogBody = document.querySelector('.el-dialog__body') as HTMLElement
    const dialogInputs = [...dialogBody.querySelectorAll('input')]
    const spiderIDInput = dialogInputs.find((input) => input.getAttribute('placeholder') === 'spider id')
    const imageInput = dialogInputs.find((input) => input.getAttribute('placeholder') === 'crawler/go-echo:latest')
    const commandInput = dialogInputs.find((input) => input.getAttribute('placeholder') === './go-echo')
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

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/projects', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/executions?project_id=project-1&limit=20&offset=0&sort_by=created_at&sort_order=desc', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      '/api/v1/spiders/spider-1/versions',
      expect.any(Object),
    )
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
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

  it('loads projects first and uses selected project for execution listing and creation', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          items: [
            {
              id: 'project-9',
              code: 'project-9',
              name: 'Project Nine',
            },
          ],
          total: 1,
          limit: 20,
          offset: 0,
        }),
      })
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
            id: 'v1',
            spiderId: 'spider-9',
            version: 1,
            registryAuthRef: 'ghcr-prod',
            image: 'crawler/go:v1',
            command: ['./crawler'],
            createdAt: '2026-05-02T00:00:00Z',
          },
        ]),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => ({
          id: 'exec-9',
          projectId: 'project-9',
          spiderId: 'spider-9',
          spiderVersion: 1,
          registryAuthRef: 'ghcr-prod',
          image: 'crawler/go:v1',
          command: ['./crawler'],
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

    const dialogBody = document.querySelector('.el-dialog__body') as HTMLElement
    const dialogInputs = [...dialogBody.querySelectorAll('input')]
    const spiderIDInput = dialogInputs.find((input) => input.getAttribute('placeholder') === 'spider id')
    ;(spiderIDInput as HTMLInputElement).value = 'spider-9'
    spiderIDInput?.dispatchEvent(new Event('input'))
    await flushPromises()

    const loadVersionsButton = [...document.querySelectorAll('button')].find((button) => button.textContent?.includes('Load Spider Versions'))
    ;(loadVersionsButton as HTMLButtonElement).click()
    await flushPromises()

    const confirmButton = [...document.querySelectorAll('.el-dialog__footer button')].find((button) =>
      button.textContent?.includes('Create')
    )
    ;(confirmButton as HTMLButtonElement).click()
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/v1/projects', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/v1/executions?project_id=project-9&limit=20&offset=0&sort_by=created_at&sort_order=desc', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(3, '/api/v1/spiders/spider-9/versions', expect.any(Object))
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      '/api/v1/executions',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          projectId: 'project-9',
          spiderId: 'spider-9',
          spiderVersion: 1,
          registryAuthRef: 'ghcr-prod',
          image: 'crawler/go:v1',
          command: ['./crawler'],
        }),
      }),
    )
  })

  it('applies node id filter when loading executions', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          items: [{ id: 'project-1', code: 'project-1', name: 'Project One' }],
          total: 1,
          limit: 20,
          offset: 0,
        }),
      })
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
        json: async () => ({
          items: [],
          total: 0,
          limit: 20,
          offset: 0,
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

    const nodeInput = container.querySelector('input[placeholder="node id"]') as HTMLInputElement
    nodeInput.value = 'node-a'
    nodeInput.dispatchEvent(new Event('input'))
    await flushPromises()

    const applyButton = [...container.querySelectorAll('button')].find((button) => button.textContent?.includes('Apply Filters'))
    ;(applyButton as HTMLButtonElement).click()
    await flushPromises()

    expect(fetchMock).toHaveBeenNthCalledWith(3, '/api/v1/executions?project_id=project-1&limit=20&offset=0&sort_by=created_at&sort_order=desc&node_id=node-a', expect.any(Object))
  })
})
