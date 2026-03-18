/**
 * Get color for bind type
 * @param bindType - SMPP bind type
 * @returns Color string
 */
export function getBindTypeColor(bindType: string): string {
  const colors: Record<string, string> = {
    TX: '#409eff',
    RX: '#67c23a',
    TR: '#e6a23c'
  }
  return colors[bindType] || '#909399'
}

/**
 * Get display text for bind type
 * @param bindType - SMPP bind type
 * @returns Display text in Chinese
 */
export function getBindTypeText(bindType: string): string {
  const texts: Record<string, string> = {
    TX: '发送器',
    RX: '接收器',
    TR: '收发器'
  }
  return texts[bindType] || bindType
}
