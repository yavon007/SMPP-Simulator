<template>
  <div class="config-page">
    <h1 class="page-title">模拟配置</h1>

    <el-card v-loading="loading">
      <el-form :model="config" label-width="160px">
        <el-form-item label="自动响应">
          <el-switch v-model="config.auto_response" />
          <div class="form-tip">启用后将自动响应submit_sm请求</div>
        </el-form-item>

        <el-form-item label="成功率">
          <el-slider v-model="config.success_rate" :min="0" :max="100" show-input />
          <div class="form-tip">设置响应成功的百分比</div>
        </el-form-item>

        <el-form-item label="响应延迟(ms)">
          <el-input-number v-model="config.response_delay" :min="0" :max="60000" :step="100" />
          <div class="form-tip">响应submit_sm的延迟时间</div>
        </el-form-item>

        <el-form-item label="自动下发状态报告">
          <el-switch v-model="config.deliver_report" />
          <div class="form-tip">启用后将自动下发deliver_sm状态报告</div>
        </el-form-item>

        <el-form-item label="状态报告延迟(ms)">
          <el-input-number v-model="config.deliver_delay" :min="0" :max="60000" :step="100" />
          <div class="form-tip">下发状态报告的延迟时间</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">
            保存配置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Data Management -->
    <el-card class="data-card">
      <template #header>
        <span>数据管理</span>
      </template>
      <div class="data-actions">
        <div class="data-action-item">
          <span>清空消息记录</span>
          <el-button type="warning" @click="handleDeleteMessages">清空消息</el-button>
        </div>
        <div class="data-action-item">
          <span>清空连接记录</span>
          <el-button type="warning" @click="handleDeleteSessions">清空连接</el-button>
        </div>
        <div class="data-action-item">
          <span>清空所有数据</span>
          <el-button type="danger" @click="handleClearAll">清空全部</el-button>
        </div>
      </div>
    </el-card>

    <!-- Help Section -->
    <el-card class="help-card">
      <template #header>
        <span>配置说明</span>
      </template>
      <div class="help-content">
        <h4>自动响应</h4>
        <p>当开启时，SMPP服务端会自动响应客户端的submit_sm请求。关闭时，需要手动通过API触发响应。</p>

        <h4>成功率</h4>
        <p>模拟运营商返回的成功率。设置为100%时所有消息都返回成功，设置为0%时所有消息都返回失败。</p>

        <h4>响应延迟</h4>
        <p>模拟网络延迟和处理时间。设置适当的延迟可以测试客户端的超时处理能力。</p>

        <h4>自动下发状态报告</h4>
        <p>开启后，服务端会在接收消息后自动下发deliver_sm状态报告。客户端需要以receiver或transceiver模式绑定才能接收。</p>

        <h4>状态报告延迟</h4>
        <p>从接收到消息到下发状态报告的延迟时间。</p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useConfigStore, useMessageStore, useSessionStore } from '@/stores'
import { dataApi } from '@/api'

const configStore = useConfigStore()
const messageStore = useMessageStore()
const sessionStore = useSessionStore()

const config = computed(() => configStore.config)
const loading = computed(() => configStore.loading)
const saving = ref(false)

const handleSave = async () => {
  saving.value = true
  try {
    await configStore.updateConfig(config.value)
    ElMessage.success('配置已保存')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleDeleteMessages = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有消息记录吗？此操作不可恢复。',
      '确认清空',
      { type: 'warning' }
    )
    await dataApi.deleteAllMessages()
    ElMessage.success('消息已清空')
    messageStore.fetchMessages()
  } catch {}
}

const handleDeleteSessions = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有连接记录吗？此操作不可恢复。',
      '确认清空',
      { type: 'warning' }
    )
    await dataApi.deleteAllSessions()
    ElMessage.success('连接记录已清空')
    sessionStore.fetchSessions()
  } catch {}
}

const handleClearAll = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有数据（消息和连接记录）吗？此操作不可恢复。',
      '确认清空',
      { type: 'warning' }
    )
    await dataApi.clearAllData()
    ElMessage.success('所有数据已清空')
    messageStore.fetchMessages()
    sessionStore.fetchSessions()
  } catch {}
}

onMounted(() => {
  configStore.fetchConfig()
})
</script>

<style scoped>
.config-page {
  max-width: 800px;
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

.data-card {
  margin-top: 20px;
}

.data-actions {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.data-action-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #ebeef5;
}

.data-action-item:last-child {
  border-bottom: none;
}

.data-action-item span {
  color: #606266;
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
