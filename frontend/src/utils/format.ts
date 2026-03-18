/**
 * Format a time string to localized format
 * @param time - ISO time string
 * @returns Formatted time string
 */
export function formatTime(time: string): string {
  return new Date(time).toLocaleString('zh-CN')
}

/**
 * Format a time string to date only
 * @param time - ISO time string
 * @returns Formatted date string
 */
export function formatDate(time: string): string {
  return new Date(time).toLocaleDateString('zh-CN')
}
