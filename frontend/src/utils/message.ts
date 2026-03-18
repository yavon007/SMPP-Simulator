import type { MessageStatus } from '@/types'

/**
 * Get Element Plus tag type for message status
 * @param status - Message status
 * @returns Tag type for el-tag component
 */
export function getStatusType(status: MessageStatus | string): string {
  const types: Record<string, string> = {
    pending: 'warning',
    delivered: 'success',
    failed: 'danger'
  }
  return types[status] || 'info'
}

/**
 * Get display text for message status
 * @param status - Message status
 * @returns Display text in Chinese
 */
export function getStatusText(status: MessageStatus | string): string {
  const texts: Record<string, string> = {
    pending: '待处理',
    delivered: '已送达',
    failed: '失败'
  }
  return texts[status] || status
}
