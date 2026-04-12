import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ProjectsView from '../views/ProjectsView.vue'
import SpidersView from '../views/SpidersView.vue'
import ExecutionsView from '../views/ExecutionsView.vue'
import DatasourcesView from '../views/DatasourcesView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: LoginView },
    { path: '/projects', component: ProjectsView },
    { path: '/spiders', component: SpidersView },
    { path: '/executions', component: ExecutionsView },
    { path: '/datasources', component: DatasourcesView },
  ],
})

export default router
