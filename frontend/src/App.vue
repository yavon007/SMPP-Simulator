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
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
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
        <div class="auth-section" v-if="!isMobile">
          <template v-if="authStore.isAuthenticated">
            <div class="user-info">
              <el-icon><User /></el-icon>
              <span>{{ authStore.username }}</span>
            </div>
            <el-button type="danger" size="small" @click="handleLogout">退出登录</el-button>
          </template>
          <el-button v-else type="primary" size="small" @click="$router.push('/login')">登录</el-button>
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
import { DataLine, Connection, Message, Setting, Promotion, User, Expand, Cellphone } from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { wsClient } from '@/utils/websocket'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const currentRoute = computed(() => route.path)
const isMobile = ref(false)
const mobileMenuOpen = ref(false)

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
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
}

.app-container {
  height: 100%;
}

.app-aside {
  background-color: #304156;
  height: 100%;
  position: relative;
  transition: transform 0.3s ease;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  border-bottom: 1px solid #3a4758;
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
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}

.auth-section {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  padding: 0 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.user-info {
  color: #bfcbd9;
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
  background: #304156;
  z-index: 1000;
  display: flex;
  align-items: center;
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
</style>
