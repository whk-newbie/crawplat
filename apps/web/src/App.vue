<script setup lang="ts">
import { computed, provide } from 'vue'
import { useLocaleStore } from './stores/locale'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import AppLayout from './components/AppLayout.vue'

const localeStore = useLocaleStore()
const elLocale = computed(() => (localeStore.locale === 'zh-CN' ? zhCn : en))
provide('elLocale', elLocale)
</script>

<template>
  <el-config-provider :locale="elLocale">
    <AppLayout>
      <router-view />
    </AppLayout>
  </el-config-provider>
</template>
