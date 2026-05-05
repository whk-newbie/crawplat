<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
</script>

<template>
  <main>
    <h1>{{ localeStore.t('pages.projects.title') }}</h1>
    <p>{{ localeStore.t('pages.projects.placeholder') }}</p>
  </main>
</template>
