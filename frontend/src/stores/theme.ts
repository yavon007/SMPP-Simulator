import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type ThemeMode = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  const STORAGE_KEY = 'theme-preference'
  
  // 获取系统主题偏好
  const getSystemTheme = (): ThemeMode => {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return 'light'
  }
  
  // 获取存储的主题或使用系统主题
  const getStoredTheme = (): ThemeMode => {
    const stored = localStorage.getItem(STORAGE_KEY) as ThemeMode | null
    if (stored === 'light' || stored === 'dark') {
      return stored
    }
    return getSystemTheme()
  }
  
  const theme = ref<ThemeMode>(getStoredTheme())
  
  // 应用主题到 DOM
  const applyTheme = (newTheme: ThemeMode) => {
    const html = document.documentElement
    if (newTheme === 'dark') {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
  }
  
  // 切换主题
  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }
  
  // 设置特定主题
  const setTheme = (newTheme: ThemeMode) => {
    theme.value = newTheme
  }
  
  // 监听主题变化，保存到 localStorage 并应用
  watch(theme, (newTheme) => {
    localStorage.setItem(STORAGE_KEY, newTheme)
    applyTheme(newTheme)
  }, { immediate: true })
  
  // 监听系统主题变化
  if (typeof window !== 'undefined' && window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      // 只有当用户没有手动设置主题时才跟随系统
      if (!localStorage.getItem(STORAGE_KEY)) {
        theme.value = e.matches ? 'dark' : 'light'
      }
    })
  }
  
  return {
    theme,
    toggleTheme,
    setTheme
  }
})
