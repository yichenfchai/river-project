import { ElMessage } from 'element-plus'
import type { MessageOptions } from 'element-plus'

const defaults: MessageOptions = {
  duration: 2000,
  showClose: true,
  grouping: true,
}

let lastMsg = { text: '', time: 0 }

/** 去重：300ms 内相同消息不重复弹 */
function dedup(text: string): boolean {
  const now = Date.now()
  if (text === lastMsg.text && now - lastMsg.time < 300) return false
  lastMsg = { text, time: now }
  return true
}

export function useMessage() {
  return {
    success(msg: string) {
      if (!dedup(msg)) return
      ElMessage.success({ ...defaults, message: msg })
    },
    error(msg: string) {
      if (!dedup(msg)) return
      ElMessage.error({ ...defaults, message: msg })
    },
    warning(msg: string) {
      if (!dedup(msg)) return
      ElMessage.warning({ ...defaults, message: msg })
    },
    info(msg: string) {
      if (!dedup(msg)) return
      ElMessage.info({ ...defaults, message: msg })
    },
  }
}
