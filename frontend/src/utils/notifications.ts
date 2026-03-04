/**
 * 全局通知工具
 */

interface NotificationDetail {
  type: 'success' | 'error' | 'info' | 'warning'
  message: string
  duration?: number
}

/**
 * 显示通知
 */
function showNotification(type: 'success' | 'error' | 'info' | 'warning', message: string, duration?: number) {
  const detail: NotificationDetail = { type, message, duration }
  window.dispatchEvent(new CustomEvent('show-notification', { detail }))
}

/**
 * 显示成功通知
 */
export function notifySuccess(message: string, duration: number = 3000) {
  showNotification('success', message, duration)
}

/**
 * 显示错误通知
 */
export function notifyError(message: string, duration: number = 5000) {
  showNotification('error', message, duration)
}

/**
 * 显示信息通知
 */
export function notifyInfo(message: string, duration: number = 3000) {
  showNotification('info', message, duration)
}

/**
 * 显示警告通知
 */
export function notifyWarning(message: string, duration: number = 4000) {
  showNotification('warning', message, duration)
}
