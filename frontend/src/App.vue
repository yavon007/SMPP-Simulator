<template>
  <el-config-provider :locale="zhCn">
    <el-container class="app-container">
      <el-aside width="200px" class="app-aside">
        <div class="logo">
          <h2>SMPP Simulator</h2>
        </div>
        <el-menu
          :default-active="currentRoute"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
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
        <div class="auth-section">
          <template v-if="authStore.isAuthenticated">
            <div class="user-info">{{ authStore.username }}</div>
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
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DataLine, Connection, Message, Setting, Promotion } from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { wsClient } from '@/utils/websocket'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const currentRoute = computed(() => route.path)

const handleLogout = () => {
  authStore.logout()
  router.push('/')
}

onMounted(() => {
  wsClient.connect()
})

onUnmounted(() => {
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
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  border-bottom: 1px solid #3a4758;
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
}
</style>
