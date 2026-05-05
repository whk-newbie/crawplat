<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
</script>

<template>
  <el-card>
    <template #header>
      <h1>{{ localeStore.t('pages.schedules.title') }}</h1>
    </template>
    <p>{{ localeStore.t('pages.schedules.placeholder') }}</p>
  </el-card>
</template>
