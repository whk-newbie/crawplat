<script setup lang="ts">
import AppLanguageSwitcher from './AppLanguageSwitcher.vue'
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()

const navItems = [
  { to: '/login', labelKey: 'navigation.login' },
  { to: '/projects', labelKey: 'navigation.projects' },
  { to: '/spiders', labelKey: 'navigation.spiders' },
  { to: '/executions', labelKey: 'navigation.executions' },
  { to: '/monitor', labelKey: 'navigation.monitor' },
  { to: '/datasources', labelKey: 'navigation.datasources' },
]
</script>

<template>
  <div class="app-layout">
    <header class="app-header">
      <router-link class="brand" to="/projects">{{ localeStore.t('app.title') }}</router-link>
      <nav class="main-nav" :aria-label="localeStore.t('navigation.main')">
        <router-link v-for="item in navItems" :key="item.to" :to="item.to">
          {{ localeStore.t(item.labelKey) }}
        </router-link>
      </nav>
      <AppLanguageSwitcher />
    </header>
    <main class="app-content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.app-header {
  align-items: center;
  border-bottom: 1px solid #dcdfe6;
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  padding: 1rem;
}

.brand {
  color: #303133;
  font-weight: 700;
  text-decoration: none;
}

.main-nav {
  display: flex;
  flex: 1;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.main-nav a {
  color: #409eff;
  text-decoration: none;
}

.main-nav a.router-link-active {
  color: #1f5fbf;
  font-weight: 600;
}

.app-content {
  padding: 1rem;
}
</style>
