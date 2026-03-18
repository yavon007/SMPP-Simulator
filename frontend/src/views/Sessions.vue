<template>
  <div class="sessions-page">
    <h1 class="page-title">连接管理</h1>

    <el-card>
      <el-table :data="sessions" v-loading="loading" stripe>
        <el-table-column prop="id" label="会话ID" width="200" />
        <el-table-column prop="system_id" label="System ID" width="150" />
        <el-table-column prop="bind_type" label="绑定类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getBindTypeColor(row.bind_type)">
              {{ row.bind_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remote_addr" label="远程地址" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '活跃' : '已断开' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="connected_at" label="连接时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.connected_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button
              type="danger"
              link
              :disabled="row.status !== 'active'"
              @click="handleDisconnect(row)"
            >
              断开
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSessionStore } from '@/stores'
import { useWebSocketEvents } from '@/composables/useWebSocketEvents'
import { formatTime } from '@/utils/format'
import { getBindTypeColor } from '@/utils/session'
import type { Session } from '@/types'

const sessionStore = useSessionStore()

const sessions = computed(() => sessionStore.sessions)
const loading = computed(() => sessionStore.loading)

const handleDisconnect = async (session: Session) => {
  try {
    await ElMessageBox.confirm(
      `确定要断开连接 "${session.system_id}" 吗？`,
      '确认断开',
      { type: 'warning' }
    )
    await sessionStore.deleteSession(session.id)
    ElMessage.success('连接已断开')
  } catch {
    // Cancelled
  }
}

// WebSocket event handlers
useWebSocketEvents({
  onSessionConnect: (session: Session) => {
    sessionStore.addSession(session)
  },
  onSessionDisconnect: (sessionId: string) => {
    sessionStore.removeSession(sessionId)
  }
})

onMounted(async () => {
  await sessionStore.fetchSessions()
})
</script>

<style scoped>
.sessions-page {
  max-width: 1200px;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #303133;
}
</style>
