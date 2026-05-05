<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLanguageSwitcher from './AppLanguageSwitcher.vue'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
const route = useRoute()

const activeIndex = computed(() => {
  const path = route.path
  if (path.startsWith('/login')) return '/login'
  if (path.startsWith('/projects')) return '/projects'
  if (path.startsWith('/spiders')) return '/spiders'
  if (path.startsWith('/executions')) return '/executions'
  if (path.startsWith('/monitor')) return '/monitor'
  if (path.startsWith('/datasources')) return '/datasources'
  if (path.startsWith('/nodes')) return '/nodes'
  if (path.startsWith('/schedules')) return '/schedules'
  return '/projects'
})

const navItems = [
  { index: '/login', labelKey: 'navigation.login' },
  { index: '/projects', labelKey: 'navigation.projects' },
  { index: '/spiders', labelKey: 'navigation.spiders' },
  { index: '/executions', labelKey: 'navigation.executions' },
  { index: '/schedules', labelKey: 'navigation.schedules' },
  { index: '/nodes', labelKey: 'navigation.nodes' },
  { index: '/monitor', labelKey: 'navigation.monitor' },
  { index: '/datasources', labelKey: 'navigation.datasources' },
]
</script>

<template>
  <el-container class="app-layout">
    <el-header class="app-header" height="auto">
      <div class="header-left">
        <router-link class="brand" to="/projects">{{ localeStore.t('app.title') }}</router-link>
        <el-menu
          :default-active="activeIndex"
          mode="horizontal"
          router
          :ellipsis="false"
          class="main-menu"
        >
          <el-menu-item v-for="item in navItems" :key="item.index" :index="item.index">
            {{ localeStore.t(item.labelKey) }}
          </el-menu-item>
        </el-menu>
      </div>
      <AppLanguageSwitcher />
    </el-header>
    <el-main class="app-content">
      <slot />
    </el-main>
  </el-container>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.app-header {
  align-items: center;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  justify-content: space-between;
  padding: 0 1rem;
}

.header-left {
  align-items: center;
  display: flex;
  gap: 1rem;
}

.brand {
  color: var(--el-text-color-primary);
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
}

.main-menu {
  border-bottom: none;
}

.app-content {
  padding: 1rem;
}
</style>
