<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
</script>

<template>
  <el-card>
    <template #header>
      <h1>{{ localeStore.t('pages.datasources.title') }}</h1>
    </template>
    <p>{{ localeStore.t('pages.datasources.placeholder') }}</p>
  </el-card>
</template>
