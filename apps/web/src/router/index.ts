import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ProjectsView from '../views/ProjectsView.vue'
import SpidersView from '../views/SpidersView.vue'
import ExecutionsView from '../views/ExecutionsView.vue'
import ExecutionDetailView from '../views/ExecutionDetailView.vue'
import DatasourcesView from '../views/DatasourcesView.vue'
import MonitorView from '../views/MonitorView.vue'
import SchedulesView from '../views/SchedulesView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: LoginView },
    { path: '/projects', component: ProjectsView },
    { path: '/spiders', component: SpidersView },
    { path: '/executions', component: ExecutionsView },
    { path: '/executions/:id', component: ExecutionDetailView },
    { path: '/schedules', component: SchedulesView },
    { path: '/datasources', component: DatasourcesView },
    { path: '/monitor', component: MonitorView },
  ],
})

export default router
