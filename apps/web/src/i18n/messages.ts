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
      register: '注册',
      projects: '项目',
      spiders: '爬虫',
      executions: '执行',
      schedules: '调度',
      nodes: '节点',
      monitor: '监控',
      datasources: '数据源',
    },
    pages: {
      login: {
        title: '登录',
        username: '用户名',
        password: '密码',
        submit: '登录',
        toRegister: '没有账号？立即注册',
        errors: {
          usernameRequired: '请输入用户名',
          passwordRequired: '请输入密码',
          invalidCredentials: '用户名或密码错误',
        },
      },
      register: {
        title: '注册',
        username: '用户名',
        password: '密码',
        submit: '注册',
        toLogin: '已有账号？立即登录',
        errors: {
          usernameRequired: '请输入用户名',
          passwordRequired: '请输入密码',
          usernameExists: '用户名已存在',
          registrationFailed: '注册失败，请稍后重试',
        },
        success: '注册成功，请登录',
      },
      projects: {
        title: '项目',
        code: '项目编码',
        name: '项目名称',
        create: '创建项目',
        edit: '编辑项目',
        delete: '删除',
        deleteConfirm: '确定删除项目 "{name}"？删除后不可恢复。',
        deleteSuccess: '项目已删除',
        createSuccess: '项目已创建',
        updateSuccess: '项目已更新',
        empty: '暂无项目，点击上方按钮创建',
        errors: {
          loadFailed: '加载项目列表失败',
          createFailed: '创建项目失败',
          deleteFailed: '删除项目失败',
          codeRequired: '请输入项目编码',
          nameRequired: '请输入项目名称',
        },
      },
      spiders: { title: '爬虫', placeholder: '爬虫管理页面（待实现）' },
      executions: { title: '执行', placeholder: '执行管理页面（待实现）' },
      schedules: { title: '调度', placeholder: '调度管理页面（待实现）' },
      nodes: { title: '节点', placeholder: '节点管理页面（待实现）' },
      datasources: { title: '数据源', placeholder: '数据源管理页面（待实现）' },
      monitor: { title: '监控', placeholder: '监控页面（待实现）' },
      executionDetail: { title: '执行详情', placeholder: '执行详情页面（待实现）' },
    },
    errors: {
      fallback: '操作失败：{message}',
      unknown: '未知错误，请稍后重试',
      gateway: {
        missingBearerToken: '登录已失效，请重新登录。',
        invalidBearerToken: '登录凭证无效，请重新登录。',
        rateLimitExceeded: '请求过于频繁，请稍后重试。',
        upstreamServiceUnavailable: '服务暂不可用，请稍后重试。',
        notFound: '请求的资源不存在。',
        badRequest: '请求参数有误。',
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
      register: 'Register',
      projects: 'Projects',
      spiders: 'Spiders',
      executions: 'Executions',
      schedules: 'Schedules',
      nodes: 'Nodes',
      monitor: 'Monitor',
      datasources: 'Datasources',
    },
    pages: {
      login: {
        title: 'Login',
        username: 'Username',
        password: 'Password',
        submit: 'Sign In',
        toRegister: "Don't have an account? Sign up",
        errors: {
          usernameRequired: 'Please enter your username',
          passwordRequired: 'Please enter your password',
          invalidCredentials: 'Invalid username or password',
        },
      },
      register: {
        title: 'Register',
        username: 'Username',
        password: 'Password',
        submit: 'Sign Up',
        toLogin: 'Already have an account? Sign in',
        errors: {
          usernameRequired: 'Please enter your username',
          passwordRequired: 'Please enter your password',
          usernameExists: 'Username already exists',
          registrationFailed: 'Registration failed. Please try again later.',
        },
        success: 'Registration successful. Please sign in.',
      },
      projects: {
        title: 'Projects',
        code: 'Code',
        name: 'Name',
        create: 'Create Project',
        edit: 'Edit Project',
        delete: 'Delete',
        deleteConfirm: 'Delete project "{name}"? This action cannot be undone.',
        deleteSuccess: 'Project deleted',
        createSuccess: 'Project created',
        updateSuccess: 'Project updated',
        empty: 'No projects yet. Click the button above to create one.',
        errors: {
          loadFailed: 'Failed to load projects',
          createFailed: 'Failed to create project',
          deleteFailed: 'Failed to delete project',
          codeRequired: 'Please enter a project code',
          nameRequired: 'Please enter a project name',
        },
      },
      spiders: { title: 'Spiders', placeholder: 'Spider management (to be implemented)' },
      executions: { title: 'Executions', placeholder: 'Execution management (to be implemented)' },
      schedules: { title: 'Schedules', placeholder: 'Schedule management (to be implemented)' },
      nodes: { title: 'Nodes', placeholder: 'Node management (to be implemented)' },
      datasources: { title: 'Datasources', placeholder: 'Datasource management (to be implemented)' },
      monitor: { title: 'Monitor', placeholder: 'Monitor (to be implemented)' },
      executionDetail: { title: 'Execution Detail', placeholder: 'Execution detail (to be implemented)' },
    },
    errors: {
      fallback: 'Operation failed: {message}',
      unknown: 'An unknown error occurred. Please try again later.',
      gateway: {
        missingBearerToken: 'Your session has expired. Please sign in again.',
        invalidBearerToken: 'Your session token is invalid. Please sign in again.',
        rateLimitExceeded: 'Too many requests. Please try again later.',
        upstreamServiceUnavailable: 'The service is temporarily unavailable. Please try again later.',
        notFound: 'The requested resource was not found.',
        badRequest: 'Invalid request parameters.',
      },
    },
  },
}

export function isSupportedLocale(value: string | null | undefined): value is Locale {
  return supportedLocales.includes(value as Locale)
}
