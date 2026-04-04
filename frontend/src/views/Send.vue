<template>
  <div class="send-page">
    <h1 class="page-title">发送消息</h1>

    <el-row :gutter="20">
      <el-col :xs="24" :sm="12">
        <el-card>
          <template #header>
            <span>发送设置</span>
          </template>
          <el-form :model="form" label-position="top" :rules="rules" ref="formRef">
            <el-form-item label="接收方连接" prop="session_id">
              <el-select
                v-model="form.session_id"
                placeholder="请选择接收方连接"
                style="width: 100%"
                @change="handleSessionChange"
              >
                <el-option
                  v-for="receiver in receivers"
                  :key="receiver.id"
                  :label="`${receiver.system_id} (${receiver.bind_type})`"
                  :value="receiver.id"
                />
              </el-select>
              <div class="form-tip">选择一个 receiver 或 transceiver 模式的连接</div>
            </el-form-item>

            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="源地址" prop="source_addr">
                  <el-input v-model="form.source_addr" placeholder="发送方号码" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="目标地址" prop="dest_addr">
                  <el-input v-model="form.dest_addr" placeholder="接收方号码" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="编码方式">
              <el-radio-group v-model="form.encoding" class="encoding-radio">
                <el-radio value="GSM7">GSM7</el-radio>
                <el-radio value="UCS2">UCS2 (中文)</el-radio>
              </el-radio-group>
              <div class="form-tip">UCS2 支持中文，GSM7 仅支持 ASCII</div>
            </el-form-item>

            <el-form-item label="消息内容" prop="content">
              <el-input
                v-model="form.content"
                type="textarea"
                :rows="3"
                placeholder="请输入消息内容"
                show-word-limit
                :maxlength="form.encoding === 'GSM7' ? 160 : 70"
              />
              <div class="form-tip">
                {{ form.encoding === 'GSM7' ? 'GSM7 最大 160 字符' : 'UCS2 最大 70 字符' }}
              </div>
            </el-form-item>

            <el-form-item class="form-actions">
              <el-button type="primary" @click="handleSend" :loading="sending" :disabled="!form.session_id">
                发送消息
              </el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12">
        <el-card class="receivers-card">
          <template #header>
            <div class="card-header">
              <span>可用连接</span>
              <el-button type="primary" link @click="fetchReceivers">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>

          <!-- 桌面端表格 -->
          <el-table :data="receivers" v-loading="loading" empty-text="暂无可用连接" class="desktop-table">
            <el-table-column prop="system_id" label="System ID" />
            <el-table-column prop="bind_type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag :type="row.bind_type === 'transceiver' ? 'success' : 'info'" size="small">
                  {{ row.bind_type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remote_addr" label="远程地址" />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button type="primary" link @click="selectReceiver(row)">选择</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 移动端卡片列表 -->
          <div class="mobile-receiver-list" v-loading="loading">
            <div class="receiver-item" v-for="receiver in receivers" :key="receiver.id" @click="selectReceiver(receiver)">
              <div class="receiver-info">
                <div class="receiver-name">{{ receiver.system_id }}</div>
                <div class="receiver-addr">{{ receiver.remote_addr }}</div>
              </div>
              <div class="receiver-meta">
                <el-tag :type="receiver.bind_type === 'transceiver' ? 'success' : 'info'" size="small">
                  {{ receiver.bind_type }}
                </el-tag>
                <el-icon class="select-icon"><Right /></el-icon>
              </div>
            </div>
            <el-empty v-if="!loading && receivers.length === 0" description="暂无可用连接" :image-size="60" />
          </div>
        </el-card>

        <el-card class="help-card">
          <template #header>
            <span>使用说明</span>
          </template>
          <div class="help-content">
            <div class="help-item">
              <strong>发送消息功能</strong>
              <p>管理员可以主动向已连接的客户端发送消息，用于测试客户端的接收功能。</p>
            </div>
            <div class="help-item">
              <strong>接收方连接</strong>
              <p>只有 receiver 或 transceiver 模式绑定的连接才能接收消息。</p>
            </div>
            <div class="help-item">
              <strong>编码方式</strong>
              <p>GSM7 适用于英文，最大 160 字符。UCS2 支持中文，最大 70 字符。</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Refresh, Right } from '@element-plus/icons-vue'
import { sendApi, type Receiver } from '@/api'

const formRef = ref<FormInstance>()
const loading = ref(false)
const sending = ref(false)
const receivers = ref<Receiver[]>([])

const form = ref({
  session_id: '',
  source_addr: '',
  dest_addr: '',
  content: '',
  encoding: 'GSM7' as 'GSM7' | 'UCS2'
})

const rules: FormRules = {
  session_id: [{ required: true, message: '请选择接收方连接', trigger: 'change' }],
  source_addr: [{ required: true, message: '请输入源地址', trigger: 'blur' }],
  dest_addr: [{ required: true, message: '请输入目标地址', trigger: 'blur' }],
  content: [{ required: true, message: '请输入消息内容', trigger: 'blur' }]
}

const fetchReceivers = async () => {
  loading.value = true
  try {
    const { data } = await sendApi.listReceivers()
    receivers.value = data.data || []
  } catch {
    ElMessage.error('获取连接列表失败')
  } finally {
    loading.value = false
  }
}

const handleSessionChange = () => {
  // 可以在这里自动填充一些默认值
}

const selectReceiver = (receiver: Receiver) => {
  form.value.session_id = receiver.id
  ElMessage.success(`已选择: ${receiver.system_id}`)
}

const handleSend = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    sending.value = true
    try {
      await sendApi.sendMessage({
        session_id: form.value.session_id,
        source_addr: form.value.source_addr,
        dest_addr: form.value.dest_addr,
        content: form.value.content,
        encoding: form.value.encoding
      })
      ElMessage.success('消息发送成功')
      form.value.content = ''
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } }
      ElMessage.error(err.response?.data?.error || '发送失败')
    } finally {
      sending.value = false
    }
  })
}

const handleReset = () => {
  formRef.value?.resetFields()
}

onMounted(() => {
  fetchReceivers()
})
</script>

<style scoped>
.send-page {
  max-width: 1200px;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #303133;
}

.encoding-radio {
  display: flex;
  gap: 16px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.form-actions {
  margin-bottom: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.receivers-card {
  margin-bottom: 20px;
}

.help-card {
  margin-top: 0;
}

.help-content {
  font-size: 14px;
}

.help-item {
  margin-bottom: 12px;
}

.help-item:last-child {
  margin-bottom: 0;
}

.help-item strong {
  color: #303133;
  display: block;
  margin-bottom: 4px;
}

.help-item p {
  margin: 0;
  color: #606266;
  line-height: 1.5;
  font-size: 13px;
}

/* 移动端接收方列表 */
.mobile-receiver-list {
  display: none;
}

.receiver-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #EBEEF5;
  cursor: pointer;
}

.receiver-item:last-child {
  border-bottom: none;
}

.receiver-item:active {
  background-color: #f5f7fa;
}

.receiver-info {
  flex: 1;
  min-width: 0;
}

.receiver-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.receiver-addr {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.receiver-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.select-icon {
  color: #909399;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .page-title {
    font-size: 18px;
    margin-bottom: 16px;
  }

  .receivers-card {
    margin-bottom: 16px;
  }

  .desktop-table {
    display: none;
  }

  .mobile-receiver-list {
    display: block;
  }

  .help-content {
    font-size: 13px;
  }

  .encoding-radio {
    flex-direction: column;
    gap: 8px;
  }
}
</style>
