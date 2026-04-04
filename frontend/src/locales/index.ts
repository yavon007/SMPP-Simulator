import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

// 获取存储的语言或浏览器语言
function getDefaultLocale(): string {
  // 优先从 localStorage 获取
  const stored = localStorage.getItem('locale')
  if (stored && (stored === 'zh-CN' || stored === 'en-US')) {
    return stored
  }
  
  // 检测浏览器语言
  const browserLang = navigator.language || navigator.languages?.[0]
  if (browserLang?.startsWith('zh')) {
    return 'zh-CN'
  }
  
  return 'en-US'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getDefaultLocale(),
  fallbackLocale: 'en-US',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

export default i18n

// 导出切换语言的工具函数
export function setLocale(locale: string): void {
  if (locale === 'zh-CN' || locale === 'en-US') {
    localStorage.setItem('locale', locale)
    i18n.global.locale.value = locale
  }
}

export function getLocale(): string {
  return i18n.global.locale.value
}
