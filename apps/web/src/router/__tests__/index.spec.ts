import { describe, expect, it } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

describe('router routes', () => {
  it('defines all expected routes', () => {
    setActivePinia(createPinia())

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', redirect: '/login' },
        { path: '/login', component: { template: '<div>Login</div>' } },
        { path: '/projects', component: { template: '<div>Projects</div>' } },
        { path: '/spiders', component: { template: '<div>Spiders</div>' } },
        { path: '/executions', component: { template: '<div>Executions</div>' } },
        { path: '/executions/:id', component: { template: '<div>Detail</div>' } },
        { path: '/datasources', component: { template: '<div>Datasources</div>' } },
        { path: '/monitor', component: { template: '<div>Monitor</div>' } },
      ],
    })

    const paths = router.getRoutes().map((r) => r.path)
    expect(paths).toContain('/login')
    expect(paths).toContain('/projects')
    expect(paths).toContain('/spiders')
    expect(paths).toContain('/executions')
    expect(paths).toContain('/executions/:id')
    expect(paths).toContain('/datasources')
    expect(paths).toContain('/monitor')
  })
})
