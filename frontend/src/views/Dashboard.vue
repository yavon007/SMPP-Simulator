<template>
  <div class="dashboard">
    <h1 class="page-title">仪表盘</h1>

    <!-- Stats Cards -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-icon connections">
            <el-icon><Connection /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.active_connections }}</div>
            <div class="stat-label">活跃连接</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-icon messages">
            <el-icon><Message /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total_messages }}</div>
            <div class="stat-label">消息总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-icon pending">
            <el-icon><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending_messages }}</div>
            <div class="stat-label">待处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-icon delivered">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.delivered_messages }}</div>
            <div class="stat-label">已送达</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Recent Messages -->
    <el-card class="recent-messages">
      <template #header>
        <div class="card-header">
          <span>最近消息</span>
          <el-button type="primary" link @click="$router.push('/messages')">
            查看全部
          </el-button>
        </div>
      </template>
      
      <!-- 桌面端表格 -->
      <el-table :data="recentMessages" v-loading="loading" stripe class="desktop-table">
        <el-table-column prop="message_id" label="消息ID" width="180" />
        <el-table-column prop="source_addr" label="发送方" width="120" />
        <el-table-column prop="dest_addr" label="接收方" width="120" />
        <el-table-column prop="content" label="内容" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 移动端卡片列表 -->
      <div class="mobile-message-list" v-loading="loading">
        <div class="message-item" v-for="msg in recentMessages" :key="msg.id">
          <div class="message-header">
            <span class="message-id">{{ msg.message_id }}</span>
            <el-tag :type="getStatusType(msg.status)" size="small">
              {{ getStatusText(msg.status) }}
            </el-tag>
          </div>
          <div class="message-body">
            <div class="message-route">
              <span class="from">{{ msg.source_addr }}</span>
              <el-icon><Right /></el-icon>
              <span class="to">{{ msg.dest_addr }}</span>
            </div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
          <div class="message-footer">
            {{ formatTime(msg.created_at) }}
          </div>
        </div>
        <el-empty v-if="!loading && recentMessages.length === 0" description="暂无消息" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { Connection, Message, Clock, CircleCheck, Right } from '@element-plus/icons-vue'
import { useStatsStore, useMessageStore } from '@/stores'
import { wsClient } from '@/utils/websocket'
import { useWebSocketEvents } from '@/composables/useWebSocketEvents'
import { formatTime } from '@/utils/format'
import { getStatusType, getStatusText } from '@/utils/message'
import type { Message as MessageType } from '@/types'

const statsStore = useStatsStore()
const messageStore = useMessageStore()

const stats = computed(() => statsStore.stats)
const recentMessages = computed(() => messageStore.messages.slice(0, 10))
const loading = computed(() => messageStore.loading)

// WebSocket event handlers
useWebSocketEvents({
  onMessageReceived: (message: MessageType) => {
    messageStore.addMessage(message)
    statsStore.updateStats({ total_messages: stats.value.total_messages + 1 })
  },
  onMessageDelivered: (messageId: string) => {
    messageStore.updateMessageStatus(messageId, 'delivered')
    statsStore.updateStats({
      pending_messages: Math.max(0, stats.value.pending_messages - 1),
      delivered_messages: stats.value.delivered_messages + 1
    })
  },
  onSessionConnect: () => {
    statsStore.updateStats({ active_connections: stats.value.active_connections + 1 })
  },
  onSessionDisconnect: () => {
    statsStore.updateStats({ active_connections: Math.max(0, stats.value.active_connections - 1) })
  }
})

onMounted(async () => {
  await Promise.all([
    statsStore.fetchStats(),
    messageStore.fetchMessages({ page_size: 10 })
  ])

  wsClient.connect()
})
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #303133;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 16px;
}

.stat-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  margin-right: 12px;
  flex-shrink: 0;
}

.stat-icon.connections {
  background: linear-gradient(135deg, #409eff, #66b1ff);
}

.stat-icon.messages {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.stat-icon.pending {
  background: linear-gradient(135deg, #e6a23c, #ebb563);
}

.stat-icon.delivered {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 13px;
  color: #909399;
}

.recent-messages {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 移动端消息列表 */
.mobile-message-list {
  display: none;
}

.message-item {
  padding: 12px 0;
  border-bottom: 1px solid #EBEEF5;
}

.message-item:last-child {
  border-bottom: none;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.message-id {
  font-size: 13px;
  color: #606266;
  font-family: monospace;
}

.message-body {
  margin-bottom: 8px;
}

.message-route {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-bottom: 6px;
}

.message-route .from {
  color: #409EFF;
}

.message-route .to {
  color: #67C23A;
}

.message-route .el-icon {
  color: #909399;
  font-size: 12px;
}

.message-content {
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-footer {
  font-size: 12px;
  color: #909399;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .page-title {
    font-size: 18px;
    margin-bottom: 16px;
  }

  .stats-row {
    margin-bottom: 16px;
  }

  .stat-card :deep(.el-card__body) {
    padding: 12px;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 20px;
    margin-right: 10px;
  }

  .stat-value {
    font-size: 20px;
  }

  .stat-label {
    font-size: 12px;
  }

  .desktop-table {
    display: none;
  }

  .mobile-message-list {
    display: block;
  }
}
</style>
