import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { defaultLocale, supportedLocales, type Locale } from '../i18n/messages'
import { normalizeLocale, translate } from '../i18n/translate'

const storageKey = 'crawler_platform_locale'

function readInitialLocale(): Locale {
  if (typeof localStorage === 'undefined') {
    return defaultLocale
  }
  return normalizeLocale(localStorage.getItem(storageKey))
}

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<Locale>(readInitialLocale())
  const availableLocales = computed(() => supportedLocales)

  function setLocale(nextLocale: Locale) {
    locale.value = normalizeLocale(nextLocale)
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(storageKey, locale.value)
    }
  }

  function t(path: string, params?: Record<string, string | number>) {
    return translate(path, locale.value, params)
  }

  return {
    availableLocales,
    locale,
    setLocale,
    t,
  }
})
