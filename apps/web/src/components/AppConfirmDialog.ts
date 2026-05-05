import { ElMessageBox } from 'element-plus'

type ConfirmOptions = {
  titleKey?: string
  title?: string
  messageKey?: string
  message?: string
  confirmButtonTextKey?: string
  confirmButtonText?: string
  cancelButtonTextKey?: string
  cancelButtonText?: string
  type?: 'warning' | 'info' | 'error'
}

export async function confirmAction(options: ConfirmOptions): Promise<boolean> {
  try {
    await ElMessageBox.confirm(
      options.message ?? (options.messageKey ?? 'Are you sure?'),
      options.title ?? (options.titleKey ?? 'Confirm'),
      {
        confirmButtonText: options.confirmButtonText ?? (options.confirmButtonTextKey ?? 'OK'),
        cancelButtonText: options.cancelButtonText ?? (options.cancelButtonTextKey ?? 'Cancel'),
        type: options.type ?? 'warning',
      },
    )
    return true
  } catch {
    return false
  }
}

export async function confirmDelete(messageKey?: string, message?: string): Promise<boolean> {
  return confirmAction({
    titleKey: 'common.actions.delete',
    messageKey: messageKey ?? 'common.actions.confirmDelete',
    message,
    confirmButtonTextKey: 'common.actions.confirm',
    cancelButtonTextKey: 'common.actions.cancel',
    type: 'warning',
  })
}
