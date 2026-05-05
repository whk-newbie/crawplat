import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useLocaleStore } from '../locale'

describe('locale store', () => {
  let storage: Record<string, string>

  beforeEach(() => {
    setActivePinia(createPinia())
    storage = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => storage[key] ?? null,
      setItem: (key: string, value: string) => {
        storage[key] = value
      },
      removeItem: (key: string) => {
        delete storage[key]
      },
      clear: () => {
        storage = {}
      },
    })
  })

  it('uses zh-CN by default', () => {
    const store = useLocaleStore()
    expect(store.locale).toBe('zh-CN')
  })

  it('persists selected locale', () => {
    const store = useLocaleStore()
    store.setLocale('en-US')

    expect(store.locale).toBe('en-US')
    expect(storage.crawler_platform_locale).toBe('en-US')
  })

  it('translates messages with current locale', () => {
    const store = useLocaleStore()
    store.setLocale('en-US')

    expect(store.t('navigation.projects')).toBe('Projects')
  })
})
