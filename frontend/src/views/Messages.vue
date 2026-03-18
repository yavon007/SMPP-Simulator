<template>
  <div class="messages-page">
    <h1 class="page-title">消息列表</h1>

    <!-- Filters -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="发送方">
          <el-input v-model="filters.source_addr" placeholder="发送方号码" clearable style="width: 140px" />
        </el-form-item>
        <el-form-item label="接收方">
          <el-input v-model="filters.dest_addr" placeholder="接收方号码" clearable style="width: 140px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 100px">
            <el-option label="全部" value="" />
            <el-option label="待处理" value="pending" />
            <el-option label="已送达" value="delivered" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 360px"
            :shortcuts="dateShortcuts"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleFilter">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
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
            <template v-if="row.status === 'pending'">
              <el-button type="success" link @click="handleDeliver(row)">
                送达
              </el-button>
              <el-button type="danger" link @click="handleFail(row)">
                失败
              </el-button>
            </template>
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
  status: '',
  source_addr: '',
  dest_addr: ''
})

const dateRange = ref<[string, string] | null>(null)

const dateShortcuts = [
  {
    text: '今天',
    value: () => {
      const start = new Date()
      start.setHours(0, 0, 0, 0)
      const end = new Date()
      return [start, end]
    }
  },
  {
    text: '最近7天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 7 * 24 * 3600 * 1000)
      return [start, end]
    }
  },
  {
    text: '最近30天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 30 * 24 * 3600 * 1000)
      return [start, end]
    }
  }
]

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

const getFilterParams = () => {
  const params: Record<string, string> = {
    status: filters.value.status,
    source_addr: filters.value.source_addr,
    dest_addr: filters.value.dest_addr
  }
  if (dateRange.value && dateRange.value.length === 2) {
    params.start_time = dateRange.value[0]
    params.end_time = dateRange.value[1]
  }
  return params
}

const handleFilter = () => {
  messageStore.fetchMessages({ ...getFilterParams(), page: 1 })
}

const handleReset = () => {
  filters.value = { status: '', source_addr: '', dest_addr: '' }
  dateRange.value = null
  messageStore.fetchMessages({ page: 1 })
}

const handlePageChange = (newPage: number) => {
  messageStore.fetchMessages({ ...getFilterParams(), page: newPage })
}

const handleSizeChange = (newSize: number) => {
  messageStore.fetchMessages({ ...getFilterParams(), page: 1, page_size: newSize })
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

const handleFail = async (row: any) => {
  try {
    await messageApi.fail(row.id)
    ElMessage.success('已标记为失败')
    messageStore.updateMessageStatus(row.id, 'failed')
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleWsMessageReceived = (data: any) => {
  // Only add if matches current filters
  const msg = data.message
  if (filters.value.status && filters.value.status !== msg.status) return
  if (filters.value.source_addr && !msg.source_addr.includes(filters.value.source_addr)) return
  if (filters.value.dest_addr && !msg.dest_addr.includes(filters.value.dest_addr)) return
  messageStore.addMessage(msg)
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
