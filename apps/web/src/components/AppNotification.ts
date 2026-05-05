import { ElMessage } from 'element-plus'

type NotificationOptions = {
  messageKey?: string
  message?: string
  duration?: number
}

export function notifySuccess(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.success({
    message: opts.message ?? '',
    duration: opts.duration ?? 3000,
  })
}

export function notifyError(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.error({
    message: opts.message ?? '',
    duration: opts.duration ?? 5000,
  })
}

export function notifyWarning(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.warning({
    message: opts.message ?? '',
    duration: opts.duration ?? 4000,
  })
}

export function notifyInfo(options: NotificationOptions | string) {
  const opts = typeof options === 'string' ? { message: options } : options
  ElMessage.info({
    message: opts.message ?? '',
    duration: opts.duration ?? 3000,
  })
}
