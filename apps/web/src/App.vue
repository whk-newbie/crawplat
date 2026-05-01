<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const isLoginPage = computed(() => route.path === '/login')

const menuItems = [
  { path: '/projects', label: 'Projects' },
  { path: '/spiders', label: 'Spiders' },
  { path: '/schedules', label: 'Schedules' },
  { path: '/executions', label: 'Executions' },
  { path: '/monitor', label: 'Monitor' },
  { path: '/nodes', label: 'Nodes' },
  { path: '/datasources', label: 'Datasources' },
]

const activeMenu = computed(() => (route.path.startsWith('/executions/') ? '/executions' : route.path))
</script>

<template>
  <div v-if="isLoginPage" class="login-shell">
    <router-view />
  </div>
  <el-container v-else class="layout-shell">
    <el-header class="layout-header">
      <div class="layout-title">Crawler Platform</div>
    </el-header>
    <el-container class="layout-body">
      <el-aside class="layout-aside">
        <el-menu :default-active="activeMenu" router>
          <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
            {{ item.label }}
          </el-menu-item>
          <el-menu-item index="/login">Login</el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-shell {
  min-height: 100vh;
}

.layout-header {
  display: flex;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid var(--el-border-color);
  background: var(--el-bg-color-page);
}

.layout-title {
  font-size: 16px;
  font-weight: 600;
}

.layout-body {
  min-height: calc(100vh - 60px);
}

.layout-aside {
  width: 220px;
  border-right: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
}

.layout-main {
  padding: 16px;
}

.login-shell {
  min-height: 100vh;
}

@media (max-width: 768px) {
  .layout-aside {
    width: 84px;
  }

  .layout-main {
    padding: 12px;
  }

  .layout-title {
    font-size: 14px;
  }
}
</style>
