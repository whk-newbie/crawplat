<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
</script>

<template>
  <el-card>
    <template #header>
      <h1>{{ localeStore.t('pages.projects.title') }}</h1>
    </template>
    <p>{{ localeStore.t('pages.projects.placeholder') }}</p>
  </el-card>
</template>
