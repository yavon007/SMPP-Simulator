import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { useThemeStore } from './stores/theme'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

// 初始化主题
const themeStore = useThemeStore()
// 确保主题在应用挂载前应用到 DOM
if (themeStore.theme === 'dark') {
  document.documentElement.classList.add('dark')
}

app.mount('#app')
