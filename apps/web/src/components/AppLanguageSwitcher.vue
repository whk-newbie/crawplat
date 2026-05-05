<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useLocaleStore } from '../stores/locale'
import type { Locale } from '../i18n/messages'

const localeStore = useLocaleStore()
const { availableLocales, locale } = storeToRefs(localeStore)

const options = [
  { value: 'zh-CN', label: '中文' },
  { value: 'en-US', label: 'English' },
]
</script>

<template>
  <div class="language-switcher">
    <span class="label">{{ localeStore.t('app.language') }}</span>
    <el-select
      :model-value="locale"
      size="small"
      style="width: 100px"
      @update:model-value="localeStore.setLocale($event as Locale)"
    >
      <el-option
        v-for="item in options"
        :key="item.value"
        :label="item.label"
        :value="item.value"
      />
    </el-select>
  </div>
</template>

<style scoped>
.language-switcher {
  align-items: center;
  display: inline-flex;
  gap: 0.5rem;
}
.label {
  white-space: nowrap;
}
</style>
