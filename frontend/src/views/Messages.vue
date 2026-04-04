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
        <el-form-item label="消息内容">
          <el-input v-model="filters.content" placeholder="搜索消息内容" clearable style="width: 180px" />
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
          <el-dropdown @command="handleExport" style="margin-left: 12px">
            <el-button type="success">
              导出 <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="csv">导出 CSV</el-dropdown-item>
                <el-dropdown-item command="json">导出 JSON</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Messages Table -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>消息列表</span>
          <div class="header-actions">
            <span v-if="selectedIds.length > 0" class="selected-count">
              已选 {{ selectedIds.length }} 条
            </span>
            <el-button
              type="danger"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
            >
              批量删除
            </el-button>
          </div>
        </div>
      </template>
      <el-table
        ref="tableRef"
        :data="messages"
        v-loading="loading"
        stripe
        @row-click="handleRowClick"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
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
    <el-dialog v-model="detailVisible" title="消息详情" width="650px" destroy-on-close>
      <div v-if="currentMessage" class="detail-content">
        <!-- 消息ID（可复制） -->
        <div class="detail-row">
          <span class="detail-label">消息ID</span>
          <div class="detail-value copyable" @click="copyToClipboard(currentMessage.message_id)">
            <span>{{ currentMessage.message_id }}</span>
            <el-icon class="copy-icon"><DocumentCopy /></el-icon>
          </div>
        </div>

        <!-- 发送方/接收方 -->
        <div class="detail-row">
          <span class="detail-label">发送方</span>
          <span class="detail-value">{{ currentMessage.source_addr }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">接收方</span>
          <span class="detail-value">{{ currentMessage.dest_addr }}</span>
        </div>

        <!-- 编码方式 -->
        <div class="detail-row">
          <span class="detail-label">编码方式</span>
          <span class="detail-value">{{ currentMessage.encoding }}</span>
        </div>

        <!-- 状态标签 -->
        <div class="detail-row">
          <span class="detail-label">状态</span>
          <el-tag :type="getStatusType(currentMessage.status)" size="large">
            {{ getStatusText(currentMessage.status) }}
          </el-tag>
        </div>

        <!-- 创建时间 -->
        <div class="detail-row">
          <span class="detail-label">创建时间</span>
          <span class="detail-value">{{ formatTime(currentMessage.created_at) }}</span>
        </div>

        <!-- 消息内容（完整显示，支持滚动） -->
        <div class="detail-row detail-row-content">
          <span class="detail-label">消息内容</span>
          <pre class="message-content-pre">{{ currentMessage.content }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DocumentCopy } from '@element-plus/icons-vue'
import type { ElTable } from 'element-plus'
import { useMessageStore } from '@/stores'
import { messageApi } from '@/api'
import { useWebSocketEvents } from '@/composables/useWebSocketEvents'
import { formatTime } from '@/utils/format'
import { getStatusType, getStatusText } from '@/utils/message'
import type { Message } from '@/types'

const messageStore = useMessageStore()

const tableRef = ref<InstanceType<typeof ElTable>>()
const selectedIds = ref<string[]>([])

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
  dest_addr: '',
  content: ''
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
const currentMessage = ref<Message | null>(null)

const getFilterParams = () => {
  const params: Record<string, string> = {
    status: filters.value.status,
    source_addr: filters.value.source_addr,
    dest_addr: filters.value.dest_addr,
    content: filters.value.content
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
  filters.value = { status: '', source_addr: '', dest_addr: '', content: '' }
  dateRange.value = null
  messageStore.fetchMessages({ page: 1 })
}

const handlePageChange = (newPage: number) => {
  messageStore.fetchMessages({ ...getFilterParams(), page: newPage })
}

const handleSizeChange = (newSize: number) => {
  messageStore.fetchMessages({ ...getFilterParams(), page: 1, page_size: newSize })
}

const handleDetail = (row: Message) => {
  currentMessage.value = row
  detailVisible.value = true
}

const handleRowClick = (row: Message) => {
  handleDetail(row)
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const handleDeliver = async (row: Message) => {
  try {
    await messageApi.deliver(row.id)
    ElMessage.success('已标记为送达')
    messageStore.updateMessageStatus(row.id, 'delivered')
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleFail = async (row: Message) => {
  try {
    await messageApi.fail(row.id)
    ElMessage.success('已标记为失败')
    messageStore.updateMessageStatus(row.id, 'failed')
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleSelectionChange = (selection: Message[]) => {
  selectedIds.value = selection.map(msg => msg.id)
}

const handleBatchDelete = async () => {
  if (selectedIds.value.length === 0) return

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 条消息吗？`,
      '批量删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await messageApi.batchDelete(selectedIds.value)
    ElMessage.success(`成功删除 ${selectedIds.value.length} 条消息`)
    
    // Clear selection and refresh list
    selectedIds.value = []
    tableRef.value?.clearSelection()
    messageStore.fetchMessages({ ...getFilterParams(), page: page.value })
  } catch {
    // User cancelled or API error
  }
}

// WebSocket event handlers
useWebSocketEvents({
  onMessageReceived: (msg: Message) => {
    // Only add if matches current filters
    if (filters.value.status && filters.value.status !== msg.status) return
    if (filters.value.source_addr && !msg.source_addr.includes(filters.value.source_addr)) return
    if (filters.value.dest_addr && !msg.dest_addr.includes(filters.value.dest_addr)) return
    if (filters.value.content && !msg.content.includes(filters.value.content)) return
    messageStore.addMessage(msg)
  },
  onMessageDelivered: (messageId: string) => {
    messageStore.updateMessageStatus(messageId, 'delivered')
  }
})

onMounted(async () => {
  await messageStore.fetchMessages()
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

/* Detail Dialog Styles */
.detail-content {
  padding: 10px 0;
}

.detail-row {
  display: flex;
  align-items: flex-start;
  margin-bottom: 16px;
}

.detail-row-content {
  flex-direction: column;
}

.detail-label {
  flex-shrink: 0;
  width: 80px;
  font-weight: 500;
  color: #606266;
  font-size: 14px;
}

.detail-value {
  flex: 1;
  color: #303133;
  font-size: 14px;
  word-break: break-all;
}

.copyable {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  margin-left: -8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.copyable:hover {
  background-color: #f5f7fa;
}

.copy-icon {
  color: #909399;
  font-size: 14px;
}

.copyable:hover .copy-icon {
  color: #409eff;
}

.message-content-pre {
  flex: 1;
  margin: 8px 0 0 0;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 300px;
  overflow-y: auto;
  color: #303133;
}
</style>
