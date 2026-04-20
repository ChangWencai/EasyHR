// frontend/src/composables/useMessage.ts
import { ElMessage } from 'element-plus'

export interface MessageOptions {
  showActions?: boolean   // show retry/contact-admin buttons (error only)
  duration?: number      // override default duration
}

export function useMessage() {
  function success(msg: string, opts?: MessageOptions): void {
    ElMessage.success({ message: msg, duration: opts?.duration ?? 2000, showClose: false })
  }

  function error(msg: string, opts?: MessageOptions): void {
    ElMessage.error({
      message: msg,
      duration: opts?.duration ?? 0,   // no auto-close per D-10-15
      showClose: true,
      // Note: action buttons (retry/contact-admin) handled by ErrorActions component, not inline
    })
  }

  // Close the currently visible message immediately (used before page navigation)
  function close(): void {
    ElMessage.closeAll()
  }

  function warning(msg: string, opts?: MessageOptions): void {
    ElMessage.warning({ message: msg, duration: opts?.duration ?? 3000, showClose: true })
  }

  function info(msg: string, opts?: MessageOptions): void {
    ElMessage.info({ message: msg, duration: opts?.duration ?? 2000, showClose: true })
  }

  return { success, error, warning, info, close }
}
