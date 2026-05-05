export const supportedLocales = ['zh-CN', 'en-US'] as const

export type Locale = (typeof supportedLocales)[number]

type MessageTree = Record<string, string | MessageTree>

export const defaultLocale: Locale = 'zh-CN'
export const fallbackLocale: Locale = 'en-US'

export const messages: Record<Locale, MessageTree> = {
  'zh-CN': {
    app: {
      title: 'Crawler Platform',
      language: '语言',
    },
    common: {
      actions: {
        create: '创建',
        edit: '编辑',
        delete: '删除',
        refresh: '刷新',
        retry: '重试',
        confirm: '确认',
        cancel: '取消',
      },
      status: {
        loading: '加载中',
        empty: '暂无数据',
        pending: '等待中',
        running: '运行中',
        succeeded: '成功',
        failed: '失败',
        online: '在线',
        offline: '离线',
      },
    },
    navigation: {
      login: '登录',
      projects: '项目',
      spiders: '爬虫',
      executions: '执行',
      monitor: '监控',
      datasources: '数据源',
    },
    pages: {
      login: {
        title: '登录',
        placeholder: 'Web MVP 外壳占位页。',
      },
      projects: {
        title: '项目',
        placeholder: 'Web MVP 外壳占位页。',
      },
    },
    errors: {
      fallback: '操作失败：{message}',
      gateway: {
        missingBearerToken: '登录已失效，请重新登录。',
        invalidBearerToken: '登录凭证无效，请重新登录。',
        rateLimitExceeded: '请求过于频繁，请稍后重试。',
        upstreamServiceUnavailable: '服务暂不可用，请稍后重试。',
      },
    },
  },
  'en-US': {
    app: {
      title: 'Crawler Platform',
      language: 'Language',
    },
    common: {
      actions: {
        create: 'Create',
        edit: 'Edit',
        delete: 'Delete',
        refresh: 'Refresh',
        retry: 'Retry',
        confirm: 'Confirm',
        cancel: 'Cancel',
      },
      status: {
        loading: 'Loading',
        empty: 'No data',
        pending: 'Pending',
        running: 'Running',
        succeeded: 'Succeeded',
        failed: 'Failed',
        online: 'Online',
        offline: 'Offline',
      },
    },
    navigation: {
      login: 'Login',
      projects: 'Projects',
      spiders: 'Spiders',
      executions: 'Executions',
      monitor: 'Monitor',
      datasources: 'Datasources',
    },
    pages: {
      login: {
        title: 'Login',
        placeholder: 'Web MVP shell placeholder.',
      },
      projects: {
        title: 'Projects',
        placeholder: 'Web MVP shell placeholder.',
      },
    },
    errors: {
      fallback: 'Operation failed: {message}',
      gateway: {
        missingBearerToken: 'Your session has expired. Please sign in again.',
        invalidBearerToken: 'Your session token is invalid. Please sign in again.',
        rateLimitExceeded: 'Too many requests. Please try again later.',
        upstreamServiceUnavailable: 'The service is temporarily unavailable. Please try again later.',
      },
    },
  },
}

export function isSupportedLocale(value: string | null | undefined): value is Locale {
  return supportedLocales.includes(value as Locale)
}
