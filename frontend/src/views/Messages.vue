<template>
  <div class="messages-page">
    <h1 class="page-title">消息列表</h1>

    <!-- Filters -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="handleFilter">
            <el-option label="全部" value="" />
            <el-option label="待处理" value="pending" />
            <el-option label="已送达" value="delivered" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="handleRefresh">刷新</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Messages Table -->
    <el-card>
      <el-table :data="messages" v-loading="loading" stripe>
        <el-table-column prop="message_id" label="消息ID" width="180" />
        <el-table-column prop="source_addr" label="发送方" width="120" />
        <el-table-column prop="dest_addr" label="接收方" width="120" />
        <el-table-column prop="content" label="内容" show-overflow-tooltip />
        <el-table-column prop="encoding" label="编码" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleDetail(row)">
              详情
            </el-button>
            <el-button
              type="success"
              link
              v-if="row.status === 'pending'"
              @click="handleDeliver(row)"
            >
              送达
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="detailVisible" title="消息详情" width="600px">
      <el-descriptions :column="2" border v-if="currentMessage">
        <el-descriptions-item label="消息ID">{{ currentMessage.message_id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentMessage.status)">
            {{ getStatusText(currentMessage.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="发送方">{{ currentMessage.source_addr }}</el-descriptions-item>
        <el-descriptions-item label="接收方">{{ currentMessage.dest_addr }}</el-descriptions-item>
        <el-descriptions-item label="编码">{{ currentMessage.encoding }}</el-descriptions-item>
        <el-descriptions-item label="序列号">{{ currentMessage.sequence_num }}</el-descriptions-item>
        <el-descriptions-item label="会话ID">{{ currentMessage.session_id }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(currentMessage.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">
          <div class="message-content">{{ currentMessage.content }}</div>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useMessageStore } from '@/stores'
import { messageApi } from '@/api'
import { wsClient } from '@/utils/websocket'

const messageStore = useMessageStore()

const messages = computed(() => messageStore.messages)
const total = computed(() => messageStore.total)
const page = computed({
  get: () => messageStore.page,
  set: (val) => { messageStore.page = val }
})
const pageSize = computed({
  get: () => messageStore.pageSize,
  set: (val) => { messageStore.pageSize = val }
})
const loading = computed(() => messageStore.loading)

const filters = ref({
  status: ''
})

const detailVisible = ref(false)
const currentMessage = ref<any>(null)

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'warning',
    delivered: 'success',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待处理',
    delivered: '已送达',
    failed: '失败'
  }
  return texts[status] || status
}

const formatTime = (time: string) => {
  return new Date(time).toLocaleString('zh-CN')
}

const handleFilter = () => {
  messageStore.fetchMessages({ status: filters.value.status, page: 1 })
}

const handleRefresh = () => {
  messageStore.fetchMessages({ status: filters.value.status })
}

const handlePageChange = (newPage: number) => {
  messageStore.fetchMessages({ status: filters.value.status, page: newPage })
}

const handleSizeChange = (newSize: number) => {
  messageStore.fetchMessages({ status: filters.value.status, page: 1, page_size: newSize })
}

const handleDetail = (row: any) => {
  currentMessage.value = row
  detailVisible.value = true
}

const handleDeliver = async (row: any) => {
  try {
    await messageApi.deliver(row.id)
    ElMessage.success('已标记为送达')
    messageStore.updateMessageStatus(row.id, 'delivered')
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleWsMessageReceived = (data: any) => {
  if (!filters.value.status || filters.value.status === data.message.status) {
    messageStore.addMessage(data.message)
  }
}

const handleWsMessageDelivered = (data: any) => {
  messageStore.updateMessageStatus(data.message_id, 'delivered')
}

onMounted(async () => {
  await messageStore.fetchMessages()

  // Register WebSocket event handlers
  wsClient.on('message_received', handleWsMessageReceived)
  wsClient.on('message_delivered', handleWsMessageDelivered)
})

onUnmounted(() => {
  // Unregister WebSocket event handlers
  wsClient.off('message_received', handleWsMessageReceived)
  wsClient.off('message_delivered', handleWsMessageDelivered)
})
</script>

<style scoped>
.messages-page {
  max-width: 1400px;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #303133;
}

.filter-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.message-content {
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
