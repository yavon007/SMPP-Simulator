<template>
  <div class="send-page">
    <h1 class="page-title">发送消息</h1>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>发送设置</span>
          </template>
          <el-form :model="form" label-width="120px" :rules="rules" ref="formRef">
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
                  :label="`${receiver.system_id} (${receiver.bind_type}) - ${receiver.remote_addr}`"
                  :value="receiver.id"
                />
              </el-select>
              <div class="form-tip">选择一个 receiver 或 transceiver 模式的连接</div>
            </el-form-item>

            <el-form-item label="源地址" prop="source_addr">
              <el-input v-model="form.source_addr" placeholder="发送方号码" />
              <div class="form-tip">消息显示的发送方号码</div>
            </el-form-item>

            <el-form-item label="目标地址" prop="dest_addr">
              <el-input v-model="form.dest_addr" placeholder="接收方号码" />
              <div class="form-tip">消息的接收方号码</div>
            </el-form-item>

            <el-form-item label="编码方式">
              <el-radio-group v-model="form.encoding">
                <el-radio label="GSM7">GSM7 (ASCII)</el-radio>
                <el-radio label="UCS2">UCS2 (中文)</el-radio>
              </el-radio-group>
              <div class="form-tip">UCS2 支持中文，GSM7 仅支持 ASCII 字符</div>
            </el-form-item>

            <el-form-item label="消息内容" prop="content">
              <el-input
                v-model="form.content"
                type="textarea"
                :rows="4"
                placeholder="请输入消息内容"
                show-word-limit
                :maxlength="form.encoding === 'GSM7' ? 160 : 70"
              />
              <div class="form-tip">
                {{ form.encoding === 'GSM7' ? 'GSM7 最大 160 字符' : 'UCS2 最大 70 字符' }}
              </div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleSend" :loading="sending" :disabled="!form.session_id">
                发送消息
              </el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>可用连接</span>
              <el-button type="primary" link @click="fetchReceivers">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          <el-table :data="receivers" v-loading="loading" empty-text="暂无可用连接">
            <el-table-column prop="system_id" label="System ID" />
            <el-table-column prop="bind_type" label="绑定类型">
              <template #default="{ row }">
                <el-tag :type="row.bind_type === 'transceiver' ? 'success' : 'info'">
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
        </el-card>

        <el-card class="help-card">
          <template #header>
            <span>使用说明</span>
          </template>
          <div class="help-content">
            <h4>发送消息功能</h4>
            <p>管理员可以主动向已连接的客户端发送 deliver_sm 消息，用于测试客户端的接收功能。</p>

            <h4>接收方连接</h4>
            <p>只有 receiver 或 transceiver 模式绑定的连接才能接收消息。transmitter 模式的连接只能发送，不能接收。</p>

            <h4>编码方式</h4>
            <p>GSM7 编码适用于纯英文和数字消息，最大支持 160 字符。UCS2 编码支持中文等 Unicode 字符，最大支持 70 字符。</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
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

const handleSessionChange = (sessionId: string) => {
  const receiver = receivers.value.find(r => r.id === sessionId)
  if (receiver) {
    // 可以在这里自动填充一些默认值
  }
}

const selectReceiver = (receiver: Receiver) => {
  form.value.session_id = receiver.id
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
      form.value.content = '' // 清空消息内容，保留其他设置
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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.help-card {
  margin-top: 20px;
}

.help-content h4 {
  margin: 16px 0 8px;
  color: #303133;
}

.help-content h4:first-child {
  margin-top: 0;
}

.help-content p {
  margin: 0;
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
}
</style>
