import { describe, expect, it } from 'vitest'
import { normalizeLocale, translate } from './translate'

describe('translate', () => {
  it('returns translated text for the active locale', () => {
    expect(translate('navigation.login', 'zh-CN')).toBe('登录')
    expect(translate('navigation.login', 'en-US')).toBe('Login')
  })

  it('falls back to the key when the message is missing', () => {
    expect(translate('missing.key', 'zh-CN')).toBe('missing.key')
  })

  it('interpolates placeholder params', () => {
    expect(translate('errors.fallback', 'zh-CN', { message: '未授权' })).toBe('操作失败：未授权')
  })
})

describe('normalizeLocale', () => {
  it('returns default locale for unsupported values', () => {
    expect(normalizeLocale('fr-FR')).toBe('zh-CN')
    expect(normalizeLocale(null)).toBe('zh-CN')
  })
})
