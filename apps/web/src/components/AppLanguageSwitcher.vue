<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLocaleStore } from '../stores/locale'
import type { Locale } from '../i18n/messages'

const localeStore = useLocaleStore()
const { availableLocales, locale } = storeToRefs(localeStore)

function onChange(event: Event) {
  const target = event.target as HTMLSelectElement
  localeStore.setLocale(target.value as Locale)
}
</script>

<template>
  <label class="language-switcher">
    <span>{{ localeStore.t('app.language') }}</span>
    <select :value="locale" @change="onChange">
      <option v-for="item in availableLocales" :key="item" :value="item">
        {{ item }}
      </option>
    </select>
  </label>
</template>

<style scoped>
.language-switcher {
  align-items: center;
  display: inline-flex;
  gap: 0.5rem;
}
</style>
