<template>
  <el-config-provider :locale="zhCn">
    <el-container class="app-container">
      <!-- 移动端顶部导航栏 -->
      <div class="mobile-header" v-if="isMobile">
        <div class="mobile-header-content">
          <el-icon class="menu-toggle" @click="mobileMenuOpen = true">
            <Expand />
          </el-icon>
          <h1 class="mobile-title">SMPP Simulator</h1>
          <div class="mobile-user">
            <el-icon class="theme-toggle-mobile" @click="themeStore.toggleTheme()">
              <Sunny v-if="themeStore.theme === 'dark'" />
              <Moon v-else />
            </el-icon>
            <template v-if="authStore.isAuthenticated">
              <el-dropdown trigger="click">
                <span class="user-dropdown">
                  <el-icon><User /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item disabled>{{ authStore.username }}</el-dropdown-item>
                    <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
            <el-button v-else type="primary" size="small" @click="$router.push('/login')">登录</el-button>
          </div>
        </div>
      </div>

      <!-- 移动端遮罩层 -->
      <div 
        class="mobile-overlay" 
        v-if="isMobile && mobileMenuOpen" 
        @click="mobileMenuOpen = false"
      ></div>

      <!-- 侧边栏 -->
      <el-aside 
        :width="isMobile ? '220px' : '200px'" 
        class="app-aside"
        :class="{ 'mobile-aside': isMobile, 'mobile-aside-open': isMobile && mobileMenuOpen }"
      >
        <div class="logo">
          <el-icon class="logo-icon"><Cellphone /></el-icon>
          <h2>SMPP Simulator</h2>
        </div>
        <el-menu
          :default-active="currentRoute"
          router
          :background-color="menuBgColor"
          :text-color="menuTextColor"
          :active-text-color="menuActiveColor"
          @select="isMobile && (mobileMenuOpen = false)"
        >
          <el-menu-item index="/">
            <el-icon><DataLine /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          <el-menu-item index="/messages">
            <el-icon><Message /></el-icon>
            <span>消息列表</span>
          </el-menu-item>
          <el-menu-item v-if="authStore.isAuthenticated" index="/sessions">
            <el-icon><Connection /></el-icon>
            <span>连接管理</span>
          </el-menu-item>
          <el-menu-item v-if="authStore.isAuthenticated" index="/send">
            <el-icon><Promotion /></el-icon>
            <span>发送消息</span>
          </el-menu-item>
          <el-menu-item v-if="authStore.isAuthenticated" index="/config">
            <el-icon><Setting /></el-icon>
            <span>模拟配置</span>
          </el-menu-item>
        </el-menu>
        <div class="sidebar-footer" v-if="!isMobile">
          <!-- 主题切换按钮 -->
          <div class="theme-toggle" @click="themeStore.toggleTheme()">
            <el-icon>
              <Sunny v-if="themeStore.theme === 'dark'" />
              <Moon v-else />
            </el-icon>
            <span>{{ themeStore.theme === 'dark' ? '浅色模式' : '深色模式' }}</span>
          </div>
          <!-- 用户认证区域 -->
          <div class="auth-section">
            <template v-if="authStore.isAuthenticated">
              <div class="user-info">
                <el-icon><User /></el-icon>
                <span>{{ authStore.username }}</span>
              </div>
              <el-button type="danger" size="small" @click="handleLogout">退出登录</el-button>
            </template>
            <el-button v-else type="primary" size="small" @click="$router.push('/login')">登录</el-button>
          </div>
        </div>
      </el-aside>

      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DataLine, Connection, Message, Setting, Promotion, User, Expand, Cellphone, Sunny, Moon } from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { wsClient } from '@/utils/websocket'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const themeStore = useThemeStore()

const currentRoute = computed(() => route.path)
const isMobile = ref(false)
const mobileMenuOpen = ref(false)

// 根据主题动态计算菜单颜色
const menuBgColor = computed(() => themeStore.theme === 'dark' ? '#1f2937' : '#304156')
const menuTextColor = computed(() => themeStore.theme === 'dark' ? '#d1d5db' : '#bfcbd9')
const menuActiveColor = computed(() => '#409EFF')

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

const handleLogout = () => {
  authStore.logout()
  router.push('/')
  mobileMenuOpen.value = false
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  wsClient.connect()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  wsClient.removeAllHandlers()
  wsClient.disconnect()
})
</script>

<style>
/* CSS 变量定义 */
:root {
  --bg-color: #f0f2f5;
  --bg-card: #ffffff;
  --text-color: #303133;
  --text-secondary: #606266;
  --border-color: #dcdfe6;
  --aside-bg: #304156;
  --aside-text: #bfcbd9;
  --aside-border: #3a4758;
  --shadow-color: rgba(0, 0, 0, 0.1);
  --hover-bg: #f5f7fa;
}

html.dark {
  --bg-color: #111827;
  --bg-card: #1f2937;
  --text-color: #e5e7eb;
  --text-secondary: #9ca3af;
  --border-color: #374151;
  --aside-bg: #1f2937;
  --aside-text: #d1d5db;
  --aside-border: #374151;
  --shadow-color: rgba(0, 0, 0, 0.3);
  --hover-bg: #374151;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
}

body {
  background-color: var(--bg-color);
  color: var(--text-color);
  transition: background-color 0.3s ease, color 0.3s ease;
}

.app-container {
  height: 100%;
}

.app-aside {
  background-color: var(--aside-bg);
  height: 100%;
  position: relative;
  transition: transform 0.3s ease, background-color 0.3s ease;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  border-bottom: 1px solid var(--aside-border);
  gap: 8px;
}

.logo-icon {
  font-size: 24px;
  color: #409EFF;
}

.logo h2 {
  font-size: 16px;
  font-weight: 600;
}

.app-main {
  background-color: var(--bg-color);
  padding: 20px;
  overflow-y: auto;
  transition: background-color 0.3s ease;
}

.sidebar-footer {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 0 0 20px 0;
}

.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  color: var(--aside-text);
  cursor: pointer;
  transition: background-color 0.2s ease;
  font-size: 14px;
}

.theme-toggle:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.auth-section {
  padding: 0 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.user-info {
  color: var(--aside-text);
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 移动端样式 */
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 50px;
  background: var(--aside-bg);
  z-index: 1000;
  display: flex;
  align-items: center;
  transition: background-color 0.3s ease;
}

.mobile-header-content {
  width: 100%;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.menu-toggle {
  font-size: 24px;
  color: #fff;
  cursor: pointer;
}

.mobile-title {
  color: #fff;
  font-size: 16px;
  font-weight: 600;
}

.mobile-user {
  display: flex;
  align-items: center;
  gap: 12px;
}

.theme-toggle-mobile {
  font-size: 20px;
  color: #fff;
  cursor: pointer;
  transition: transform 0.2s ease;
}

.theme-toggle-mobile:hover {
  transform: scale(1.1);
}

.user-dropdown {
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1001;
}

.mobile-aside {
  position: fixed;
  top: 0;
  left: 0;
  height: 100%;
  z-index: 1002;
  transform: translateX(-100%);
}

.mobile-aside-open {
  transform: translateX(0);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .app-container {
    flex-direction: column;
    padding-top: 50px;
  }

  .app-main {
    padding: 12px;
  }

  .logo {
    height: 50px;
  }
}

/* 深色模式下的 Element Plus 组件样式覆盖 */
html.dark .el-card {
  background-color: var(--bg-card);
  border-color: var(--border-color);
}

html.dark .el-table {
  background-color: var(--bg-card);
  --el-table-bg-color: var(--bg-card);
  --el-table-tr-bg-color: var(--bg-card);
  --el-table-header-bg-color: var(--bg-card);
  --el-table-row-hover-bg-color: var(--hover-bg);
  --el-table-border-color: var(--border-color);
  --el-table-text-color: var(--text-color);
}

html.dark .el-table th.el-table__cell {
  background-color: var(--bg-card);
}

html.dark .el-input__wrapper {
  background-color: var(--bg-card);
  box-shadow: 0 0 0 1px var(--border-color) inset;
}

html.dark .el-input__inner {
  color: var(--text-color);
}

html.dark .el-select__wrapper {
  background-color: var(--bg-card);
}

html.dark .el-form-item__label {
  color: var(--text-secondary);
}

html.dark .el-dialog {
  background-color: var(--bg-card);
}

html.dark .el-dialog__title {
  color: var(--text-color);
}

html.dark .el-pagination {
  --el-pagination-bg-color: var(--bg-card);
  --el-pagination-text-color: var(--text-color);
  --el-pagination-button-bg-color: var(--bg-card);
  --el-pagination-hover-color: #409EFF;
}

html.dark .el-descriptions {
  --el-descriptions-item-bordered-label-background: var(--hover-bg);
}

html.dark .el-empty__description {
  color: var(--text-secondary);
}

html.dark .el-statistic__head {
  color: var(--text-secondary);
}

html.dark .el-statistic__content {
  color: var(--text-color);
}

html.dark .el-dropdown-menu {
  background-color: var(--bg-card);
}

html.dark .el-dropdown-menu__item {
  color: var(--text-color);
}

html.dark .el-dropdown-menu__item:hover {
  background-color: var(--hover-bg);
}

/* 深色模式下修复硬编码颜色 */
html.dark .page-title,
html.dark .card-header,
html.dark .stat-value,
html.dark .message-content,
html.dark .receiver-name,
html.dark h1,
html.dark h2,
html.dark h3,
html.dark h4,
html.dark .help-content strong {
  color: var(--text-color) !important;
}

html.dark .stat-label,
html.dark .form-tip,
html.dark .message-footer,
html.dark .receiver-addr,
html.dark .help-content p,
html.dark .help-item p,
html.dark .config-page .form-tip,
html.dark .login-footer p {
  color: var(--text-secondary) !important;
}

html.dark .message-route .from,
html.dark .message-route .to {
  color: var(--text-color) !important;
}

/* 修复 Element Plus 组件在深色模式下的颜色 */
html.dark .el-card__header {
  color: var(--text-color);
  border-bottom-color: var(--border-color);
}

html.dark .el-message-box__title {
  color: var(--text-color);
}

html.dark .el-message-box__content {
  color: var(--text-secondary);
}

html.dark .el-radio__label {
  color: var(--text-color);
}

html.dark .el-checkbox__label {
  color: var(--text-color);
}

html.dark .el-textarea__inner {
  background-color: var(--bg-card);
  color: var(--text-color);
  box-shadow: 0 0 0 1px var(--border-color) inset;
}

html.dark .el-slider__runway {
  background-color: var(--border-color);
}

html.dark .el-switch__core {
  background-color: var(--border-color);
}

html.dark .el-divider {
  border-top-color: var(--border-color);
}

html.dark .el-tag--info {
  background-color: var(--hover-bg);
  color: var(--text-color);
  border-color: var(--border-color);
}

html.dark .el-progress__text {
  color: var(--text-color) !important;
}

html.dark .el-progress-bar__innerText {
  color: var(--text-color);
}

/* 修复表格内文字 */
html.dark .el-table .cell {
  color: var(--text-color);
}

html.dark .el-table__empty-text {
  color: var(--text-secondary);
}

/* 修复弹窗内文字 */
html.dark .el-dialog__body {
  color: var(--text-color);
}

html.dark .el-dialog__footer {
  border-top-color: var(--border-color);
}

/* 修复下拉选择框 */
html.dark .el-select-dropdown {
  background-color: var(--bg-card);
}

html.dark .el-select-dropdown__item {
  color: var(--text-color);
}

html.dark .el-select-dropdown__item.hover,
html.dark .el-select-dropdown__item:hover {
  background-color: var(--hover-bg);
}

/* 修复时间选择器 */
html.dark .el-date-picker {
  background-color: var(--bg-card);
}

html.dark .el-picker-panel__body {
  color: var(--text-color);
}

</style>
