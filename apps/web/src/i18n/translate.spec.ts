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

  it('returns new common action keys', () => {
    expect(translate('common.actions.search', 'zh-CN')).toBe('搜索')
    expect(translate('common.actions.search', 'en-US')).toBe('Search')
    expect(translate('common.actions.submit', 'zh-CN')).toBe('提交')
    expect(translate('common.actions.reset', 'zh-CN')).toBe('重置')
  })

  it('returns new error keys', () => {
    expect(translate('common.error.default', 'zh-CN')).toBe('操作失败，请稍后重试')
    expect(translate('common.error.default', 'en-US')).toBe('Operation failed, please try again later')
    expect(translate('common.error.network', 'zh-CN')).toBe('网络连接异常')
    expect(translate('common.error.unauthorized', 'en-US')).toBe('Session expired, please sign in again')
  })

  it('returns new page keys', () => {
    expect(translate('pages.executions.title', 'zh-CN')).toBe('执行管理')
    expect(translate('pages.executions.title', 'en-US')).toBe('Executions')
    expect(translate('pages.nodes.title', 'en-US')).toBe('Nodes')
    expect(translate('pages.nodes.title', 'zh-CN')).toBe('节点')
    expect(translate('pages.schedules.title', 'zh-CN')).toBe('调度')
    expect(translate('pages.schedules.title', 'en-US')).toBe('Schedules')
  })
})

describe('normalizeLocale', () => {
  it('returns default locale for unsupported values', () => {
    expect(normalizeLocale('fr-FR')).toBe('zh-CN')
    expect(normalizeLocale(null)).toBe('zh-CN')
  })
})
