<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLocaleStore } from './stores/locale'

const localeStore = useLocaleStore()
const { availableLocales, locale } = storeToRefs(localeStore)
</script>

<template>
  <div>
    <header>
      <nav aria-label="Main navigation">
        <router-link to="/login">{{ localeStore.t('navigation.login') }}</router-link>
        <router-link to="/projects">{{ localeStore.t('navigation.projects') }}</router-link>
        <router-link to="/spiders">{{ localeStore.t('navigation.spiders') }}</router-link>
        <router-link to="/executions">{{ localeStore.t('navigation.executions') }}</router-link>
        <router-link to="/monitor">{{ localeStore.t('navigation.monitor') }}</router-link>
        <router-link to="/datasources">{{ localeStore.t('navigation.datasources') }}</router-link>
      </nav>
      <label>
        {{ localeStore.t('app.language') }}
        <select :value="locale" @change="localeStore.setLocale(($event.target as HTMLSelectElement).value as 'zh-CN' | 'en-US')">
          <option v-for="item in availableLocales" :key="item" :value="item">
            {{ item }}
          </option>
        </select>
      </label>
    </header>
    <router-view />
  </div>
</template>
