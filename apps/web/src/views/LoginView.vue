<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const localeStore = useLocaleStore()
</script>

<template>
  <main>
    <h1>{{ localeStore.t('pages.login.title') }}</h1>
    <p>{{ localeStore.t('pages.login.placeholder') }}</p>
  </main>
</template>
