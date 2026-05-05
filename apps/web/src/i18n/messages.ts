export const supportedLocales = ['zh-CN', 'en-US'] as const

export type Locale = (typeof supportedLocales)[number]

type MessageTree = { [key: string]: string | MessageTree }

export const defaultLocale: Locale = 'zh-CN'
export const fallbackLocale: Locale = 'en-US'

export const messages: Record<Locale, MessageTree> = {
  'zh-CN': {
    app: {
      title: 'Crawler Platform',
      language: '语言',
    },
    common: {
      state: {
        loading: '加载中',
        empty: '暂无数据',
        error: '加载失败',
        pending: '等待中',
        running: '运行中',
        succeeded: '成功',
        failed: '失败',
        online: '在线',
        offline: '离线',
      },
      error: {
        default: '操作失败，请稍后重试',
        network: '网络连接异常',
        unauthorized: '登录已过期，请重新登录',
        forbidden: '无权限访问',
        notFound: '资源不存在',
        serverError: '服务异常',
      },
      actions: {
        create: '创建',
        edit: '编辑',
        delete: '删除',
        refresh: '刷新',
        retry: '重试',
        confirm: '确认',
        cancel: '取消',
        search: '搜索',
        reset: '重置',
        submit: '提交',
        back: '返回',
        save: '保存',
        close: '关闭',
        confirmDelete: '确定要删除吗？此操作不可撤销。',
      },
    },
    navigation: {
      main: '主导航',
      login: '登录',
      projects: '项目',
      spiders: '爬虫',
      executions: '执行',
      monitor: '监控',
      datasources: '数据源',
      nodes: '节点',
      schedules: '调度',
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
      spiders: {
        title: '爬虫管理',
        placeholder: 'Web MVP 外壳占位页。',
      },
      executions: {
        title: '执行管理',
        createTitle: '创建执行',
        detailTitle: '执行详情',
        lookupTitle: '查看执行',
        projectId: '项目 ID',
        spiderId: '爬虫 ID',
        image: '镜像',
        command: '命令',
        placeholder: 'Web MVP 外壳占位页。',
        creating: '创建中...',
        createAction: '创建执行',
        openAction: '打开',
        lookupPlaceholder: '执行 ID',
        createdMessage: '已创建执行',
      },
      monitor: {
        title: '监控',
        placeholder: 'Web MVP 外壳占位页。',
      },
      datasources: {
        title: '数据源',
        placeholder: 'Web MVP 外壳占位页。',
      },
      nodes: {
        title: '节点',
        placeholder: 'Web MVP 外壳占位页。',
      },
      schedules: {
        title: '调度',
        placeholder: 'Web MVP 外壳占位页。',
      },
      notFound: {
        title: '404',
        description: '页面不存在',
      },
    },
    errors: {
      fallback: '操作失败：{message}',
      gateway: {
        badRequest: '请求参数错误',
        missingBearerToken: '登录已失效，请重新登录。',
        invalidBearerToken: '登录凭证无效，请重新登录。',
        rateLimitExceeded: '请求过于频繁，请稍后重试。',
        notFound: '接口不存在',
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
      state: {
        loading: 'Loading',
        empty: 'No data',
        error: 'Load failed',
        pending: 'Pending',
        running: 'Running',
        succeeded: 'Succeeded',
        failed: 'Failed',
        online: 'Online',
        offline: 'Offline',
      },
      error: {
        default: 'Operation failed, please try again later',
        network: 'Network error',
        unauthorized: 'Session expired, please sign in again',
        forbidden: 'Access denied',
        notFound: 'Resource not found',
        serverError: 'Service error',
      },
      actions: {
        create: 'Create',
        edit: 'Edit',
        delete: 'Delete',
        refresh: 'Refresh',
        retry: 'Retry',
        confirm: 'Confirm',
        cancel: 'Cancel',
        search: 'Search',
        reset: 'Reset',
        submit: 'Submit',
        back: 'Back',
        save: 'Save',
        close: 'Close',
        confirmDelete: 'Are you sure? This action cannot be undone.',
      },
    },
    navigation: {
      main: 'Main Navigation',
      login: 'Login',
      projects: 'Projects',
      spiders: 'Spiders',
      executions: 'Executions',
      monitor: 'Monitor',
      datasources: 'Datasources',
      nodes: 'Nodes',
      schedules: 'Schedules',
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
      spiders: {
        title: 'Spiders',
        placeholder: 'Web MVP shell placeholder.',
      },
      executions: {
        title: 'Executions',
        createTitle: 'Create Execution',
        detailTitle: 'Execution Detail',
        lookupTitle: 'Open Execution',
        projectId: 'Project ID',
        spiderId: 'Spider ID',
        image: 'Image',
        command: 'Command',
        placeholder: 'Web MVP shell placeholder.',
        creating: 'Creating...',
        createAction: 'Create Execution',
        openAction: 'Open',
        lookupPlaceholder: 'execution id',
        createdMessage: 'Created execution',
      },
      monitor: {
        title: 'Monitor',
        placeholder: 'Web MVP shell placeholder.',
      },
      datasources: {
        title: 'Datasources',
        placeholder: 'Web MVP shell placeholder.',
      },
      nodes: {
        title: 'Nodes',
        placeholder: 'Web MVP shell placeholder.',
      },
      schedules: {
        title: 'Schedules',
        placeholder: 'Web MVP shell placeholder.',
      },
      notFound: {
        title: '404',
        description: 'Page not found',
      },
    },
    errors: {
      fallback: 'Operation failed: {message}',
      gateway: {
        badRequest: 'Bad request',
        missingBearerToken: 'Your session has expired. Please sign in again.',
        invalidBearerToken: 'Your session token is invalid. Please sign in again.',
        rateLimitExceeded: 'Too many requests. Please try again later.',
        notFound: 'API not found',
        upstreamServiceUnavailable: 'The service is temporarily unavailable. Please try again later.',
      },
    },
  },
}

export function isSupportedLocale(value: string | null | undefined): value is Locale {
  return supportedLocales.includes(value as Locale)
}
