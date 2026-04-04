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

    <!-- Message Status Distribution Chart -->
    <el-card class="status-chart-card">
      <template #header>
        <div class="card-header">
          <span>消息状态分布</span>
        </div>
      </template>
      <div class="status-chart">
        <div class="chart-item">
          <div class="chart-label">
            <span class="status-dot pending"></span>
            <span>待处理</span>
          </div>
          <el-progress 
            :percentage="pendingPercent" 
            :stroke-width="20"
            :show-text="false"
            class="progress-bar"
          />
          <div class="chart-value">
            <span class="count">{{ stats.pending_messages }}</span>
            <span class="percent">{{ pendingPercent.toFixed(1) }}%</span>
          </div>
        </div>
        <div class="chart-item">
          <div class="chart-label">
            <span class="status-dot delivered"></span>
            <span>已送达</span>
          </div>
          <el-progress 
            :percentage="deliveredPercent" 
            :stroke-width="20"
            :show-text="false"
            status="success"
            class="progress-bar"
          />
          <div class="chart-value">
            <span class="count">{{ stats.delivered_messages }}</span>
            <span class="percent">{{ deliveredPercent.toFixed(1) }}%</span>
          </div>
        </div>
        <div class="chart-item">
          <div class="chart-label">
            <span class="status-dot failed"></span>
            <span>失败</span>
          </div>
          <el-progress 
            :percentage="failedPercent" 
            :stroke-width="20"
            :show-text="false"
            status="exception"
            class="progress-bar"
          />
          <div class="chart-value">
            <span class="count">{{ stats.failed_messages }}</span>
            <span class="percent">{{ failedPercent.toFixed(1) }}%</span>
          </div>
        </div>
      </div>
      <!-- Visual pie chart using CSS -->
      <div class="pie-chart-container">
        <div class="pie-chart" :style="pieChartStyle">
          <div class="pie-center">
            <div class="pie-total">{{ stats.total_messages }}</div>
            <div class="pie-label">消息总数</div>
          </div>
        </div>
        <div class="pie-legend">
          <div class="legend-item">
            <span class="legend-color pending"></span>
            <span>待处理 {{ stats.pending_messages }}</span>
          </div>
          <div class="legend-item">
            <span class="legend-color delivered"></span>
            <span>已送达 {{ stats.delivered_messages }}</span>
          </div>
          <div class="legend-item">
            <span class="legend-color failed"></span>
            <span>失败 {{ stats.failed_messages }}</span>
          </div>
        </div>
      </div>
    </el-card>

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

// Calculate percentages for chart
const totalProcessed = computed(() => {
  return stats.value.pending_messages + stats.value.delivered_messages + stats.value.failed_messages
})

const pendingPercent = computed(() => {
  if (totalProcessed.value === 0) return 0
  return (stats.value.pending_messages / totalProcessed.value) * 100
})

const deliveredPercent = computed(() => {
  if (totalProcessed.value === 0) return 0
  return (stats.value.delivered_messages / totalProcessed.value) * 100
})

const failedPercent = computed(() => {
  if (totalProcessed.value === 0) return 0
  return (stats.value.failed_messages / totalProcessed.value) * 100
})

// Pie chart conic-gradient style
const pieChartStyle = computed(() => {
  const pending = pendingPercent.value
  const delivered = deliveredPercent.value
  const failed = failedPercent.value
  
  if (totalProcessed.value === 0) {
    return { background: '#E4E7ED' }
  }
  
  return {
    background: `conic-gradient(
      #E6A23C 0% ${pending}%,
      #67C23A ${pending}% ${pending + delivered}%,
      #F56C6C ${pending + delivered}% 100%
    )`
  }
})

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

/* Status Chart Styles */
.status-chart-card {
  margin-top: 20px;
}

.status-chart {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 24px;
}

.chart-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.chart-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.status-dot.pending {
  background-color: #E6A23C;
}

.status-dot.delivered {
  background-color: #67C23A;
}

.status-dot.failed {
  background-color: #F56C6C;
}

.progress-bar {
  width: 100%;
}

.chart-value {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.chart-value .count {
  font-size: 20px;
  font-weight: bold;
  color: #303133;
}

.chart-value .percent {
  font-size: 13px;
  color: #909399;
}

/* Pie Chart Styles */
.pie-chart-container {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 40px;
  padding: 20px 0;
}

.pie-chart {
  width: 160px;
  height: 160px;
  border-radius: 50%;
  position: relative;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.pie-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100px;
  height: 100px;
  background: #fff;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.05);
}

.pie-total {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.pie-label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.pie-legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 4px;
}

.legend-color.pending {
  background-color: #E6A23C;
}

.legend-color.delivered {
  background-color: #67C23A;
}

.legend-color.failed {
  background-color: #F56C6C;
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

  /* Chart mobile styles */
  .status-chart {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .chart-value .count {
    font-size: 18px;
  }

  .pie-chart-container {
    flex-direction: column;
    gap: 20px;
  }

  .pie-chart {
    width: 140px;
    height: 140px;
  }

  .pie-center {
    width: 80px;
    height: 80px;
  }

  .pie-total {
    font-size: 22px;
  }

  .pie-legend {
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    gap: 16px;
  }
}
</style>
