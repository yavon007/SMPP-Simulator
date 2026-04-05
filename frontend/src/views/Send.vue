<template>
  <div class="send-page">
    <h1 class="page-title">发送消息</h1>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="send-tabs">
      <!-- 下发消息 Tab -->
      <el-tab-pane label="下发消息" name="deliver">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-card>
              <template #header>
                <span>发送设置</span>
              </template>
              <el-form :model="deliverForm" label-position="top" :rules="rules" ref="deliverFormRef">
                <el-form-item label="接收方连接" prop="session_id">
                  <el-select
                    v-model="deliverForm.session_id"
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
                      <el-input v-model="deliverForm.source_addr" placeholder="发送方号码" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="目标地址" prop="dest_addr">
                      <el-input v-model="deliverForm.dest_addr" placeholder="接收方号码" />
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-form-item label="编码方式">
                  <el-radio-group v-model="deliverForm.encoding" class="encoding-radio">
                    <el-radio value="GSM7">GSM7</el-radio>
                    <el-radio value="UCS2">UCS2 (中文)</el-radio>
                  </el-radio-group>
                  <div class="form-tip">UCS2 支持中文，GSM7 仅支持 ASCII</div>
                </el-form-item>

                <el-form-item label="消息内容" prop="content">
                  <div class="content-header">
                    <el-button type="primary" link @click="showTemplateDialog">
                      <el-icon><DocumentCopy /></el-icon>
                      使用模板
                    </el-button>
                  </div>
                  <el-input
                    v-model="deliverForm.content"
                    type="textarea"
                    :rows="3"
                    placeholder="请输入消息内容"
                    show-word-limit
                    :maxlength="deliverForm.encoding === 'GSM7' ? 160 : 70"
                  />
                  <div class="form-tip">
                    {{ deliverForm.encoding === 'GSM7' ? 'GSM7 最大 160 字符' : 'UCS2 最大 70 字符' }}
                  </div>
                </el-form-item>

                <el-form-item class="form-actions">
                  <el-button type="primary" @click="handleDeliverSend" :loading="sending" :disabled="!deliverForm.session_id">
                    发送消息
                  </el-button>
                  <el-button @click="handleDeliverReset">重置</el-button>
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
                  <strong>下发消息</strong>
                  <p>向已连接的客户端发送 deliver_sm 消息，用于测试客户端的接收功能。</p>
                </div>
                <div class="help-item">
                  <strong>接收方连接</strong>
                  <p>只有 receiver 或 transceiver 模式绑定的连接才能接收消息。</p>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <!-- 主动发送 Tab -->
      <el-tab-pane label="主动发送" name="outbound">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-card>
              <template #header>
                <span>连接设置</span>
              </template>
              <el-form :model="connectForm" label-position="top" ref="connectFormRef">
                <el-row :gutter="12">
                  <el-col :span="16">
                    <el-form-item label="SMSC 地址" prop="host">
                      <el-input v-model="connectForm.host" placeholder="SMSC IP 或域名" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="端口" prop="port">
                      <el-input v-model="connectForm.port" placeholder="2775" />
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-row :gutter="12">
                  <el-col :span="12">
                    <el-form-item label="System ID" prop="system_id">
                      <el-input v-model="connectForm.system_id" placeholder="用户名" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="密码" prop="password">
                      <el-input v-model="connectForm.password" type="password" placeholder="密码" show-password />
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-form-item label="绑定类型">
                  <el-radio-group v-model="connectForm.bind_type" class="encoding-radio">
                    <el-radio value="transceiver">Transceiver</el-radio>
                    <el-radio value="transmitter">Transmitter</el-radio>
                    <el-radio value="receiver">Receiver</el-radio>
                  </el-radio-group>
                  <div class="form-tip">Transceiver 可收发，Transmitter 只能发，Receiver 只能收</div>
                </el-form-item>

                <el-form-item class="form-actions">
                  <el-button type="primary" @click="handleConnect" :loading="connecting">
                    连接
                  </el-button>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>

          <el-col :xs="24" :sm="12">
            <el-card class="outbound-sessions-card">
              <template #header>
                <div class="card-header">
                  <span>已建立的连接</span>
                  <el-button type="primary" link @click="fetchOutboundSessions">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </template>

              <el-table :data="outboundSessions" v-loading="outboundLoading" empty-text="暂无连接" class="desktop-table">
                <el-table-column prop="system_id" label="System ID" />
                <el-table-column prop="remote_addr" label="远程地址" />
                <el-table-column prop="bind_type" label="类型" width="90">
                  <template #default="{ row }">
                    <el-tag :type="row.bind_type === 'transceiver' ? 'success' : 'info'" size="small">
                      {{ row.bind_type }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'active' ? 'success' : (row.status === 'error' ? 'danger' : 'warning')" size="small">
                      {{ row.status }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="130">
                  <template #default="{ row }">
                    <el-button
                      type="primary"
                      link
                      @click="selectOutboundSession(row)"
                      :disabled="row.status !== 'active'"
                    >
                      发送
                    </el-button>
                    <el-button type="danger" link @click="handleDisconnect(row)">断开</el-button>
                  </template>
                </el-table-column>
              </el-table>

              <div class="mobile-outbound-list" v-loading="outboundLoading">
                <div class="outbound-item" v-for="session in outboundSessions" :key="session.id">
                  <div class="outbound-info" @click="selectOutboundSession(session)">
                    <div class="outbound-name">{{ session.system_id }}</div>
                    <div class="outbound-addr">{{ session.remote_addr }}</div>
                    <div class="outbound-error" v-if="session.error">{{ session.error }}</div>
                  </div>
                  <div class="outbound-meta">
                    <el-tag :type="session.status === 'active' ? 'success' : (session.status === 'error' ? 'danger' : 'warning')" size="small">
                      {{ session.status }}
                    </el-tag>
                    <el-button type="danger" link size="small" @click="handleDisconnect(session)">断开</el-button>
                  </div>
                </div>
                <el-empty v-if="!outboundLoading && outboundSessions.length === 0" description="暂无连接" :image-size="60" />
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 发送消息表单 -->
        <el-card class="outbound-send-card" v-if="selectedOutboundSession">
          <template #header>
            <div class="card-header">
              <span>发送消息 ({{ selectedOutboundSession.system_id }})</span>
              <el-button type="primary" link @click="selectedOutboundSession = null">
                关闭
              </el-button>
            </div>
          </template>
          <el-form :model="outboundSendForm" label-position="top" ref="outboundSendFormRef">
            <el-row :gutter="20">
              <el-col :xs="24" :sm="8">
                <el-form-item label="源地址" prop="source_addr">
                  <el-input v-model="outboundSendForm.source_addr" placeholder="发送方号码" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="8">
                <el-form-item label="目标地址" prop="dest_addr">
                  <el-input v-model="outboundSendForm.dest_addr" placeholder="接收方号码" />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="8">
                <el-form-item label="编码方式">
                  <el-radio-group v-model="outboundSendForm.encoding">
                    <el-radio value="GSM7">GSM7</el-radio>
                    <el-radio value="UCS2">UCS2</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="消息内容" prop="content">
              <el-input
                v-model="outboundSendForm.content"
                type="textarea"
                :rows="2"
                placeholder="请输入消息内容"
                show-word-limit
                :maxlength="outboundSendForm.encoding === 'GSM7' ? 160 : 70"
              />
            </el-form-item>

            <el-form-item class="form-actions">
              <el-button type="primary" @click="handleOutboundSend" :loading="outboundSending">
                发送消息
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="help-card" v-if="!selectedOutboundSession">
          <template #header>
            <span>使用说明</span>
          </template>
          <div class="help-content">
            <div class="help-item">
              <strong>主动发送</strong>
              <p>主动连接外部 SMSC，发送 submit_sm 消息。适用于测试 SMSC 或网关。</p>
            </div>
            <div class="help-item">
              <strong>绑定类型</strong>
              <p>Transceiver 可收发消息，Transmitter 只能发消息，Receiver 只能收消息。</p>
            </div>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- Template Selection Dialog -->
    <el-dialog v-model="templateDialogVisible" title="选择消息模板" width="600px" destroy-on-close>
      <div v-loading="templatesLoading">
        <el-table :data="templates" empty-text="暂无模板" @row-click="selectTemplate" highlight-current-row>
          <el-table-column prop="name" label="模板名称" width="120" />
          <el-table-column prop="content" label="内容" show-overflow-tooltip />
          <el-table-column prop="encoding" label="编码" width="80" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button type="primary" link @click.stop="selectTemplate(row)">选择</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="showManageDialog">管理模板</el-button>
      </template>
    </el-dialog>

    <!-- Template Management Dialog -->
    <el-dialog v-model="manageDialogVisible" title="管理消息模板" width="700px" destroy-on-close>
      <div class="template-manage-header">
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          新建模板
        </el-button>
      </div>
      <el-table :data="templates" v-loading="templatesLoading" empty-text="暂无模板">
        <el-table-column prop="name" label="模板名称" width="120" />
        <el-table-column prop="content" label="内容" show-overflow-tooltip />
        <el-table-column prop="encoding" label="编码" width="80" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button type="primary" link @click="editTemplate(row)">编辑</el-button>
            <el-button type="danger" link @click="deleteTemplate(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Create/Edit Template Dialog -->
    <el-dialog v-model="editDialogVisible" :title="editingTemplate ? '编辑模板' : '新建模板'" width="500px" destroy-on-close>
      <el-form :model="templateForm" label-position="top" ref="templateFormRef" :rules="templateRules">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="消息内容" prop="content">
          <el-input
            v-model="templateForm.content"
            type="textarea"
            :rows="4"
            placeholder="请输入模板内容，可使用 {code}、{order_id} 等占位符"
          />
        </el-form-item>
        <el-form-item label="编码方式">
          <el-radio-group v-model="templateForm.encoding">
            <el-radio value="GSM7">GSM7</el-radio>
            <el-radio value="UCS2">UCS2 (中文)</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveTemplate" :loading="savingTemplate">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Refresh, Right, DocumentCopy, Plus } from '@element-plus/icons-vue'
import { sendApi, templateApi, outboundApi, type Receiver, type MessageTemplate, type OutboundSession } from '@/api'

// Tab state
const activeTab = ref('deliver')

// Deliver form state
const deliverFormRef = ref<FormInstance>()
const loading = ref(false)
const sending = ref(false)
const receivers = ref<Receiver[]>([])

const deliverForm = ref({
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

// Outbound state
const connectFormRef = ref<FormInstance>()
const outboundLoading = ref(false)
const connecting = ref(false)
const outboundSending = ref(false)
const outboundSessions = ref<OutboundSession[]>([])
const selectedOutboundSession = ref<OutboundSession | null>(null)

const connectForm = ref({
  host: '',
  port: '2775',
  system_id: '',
  password: '',
  bind_type: 'transceiver' as 'transmitter' | 'receiver' | 'transceiver'
})

const outboundSendForm = ref({
  source_addr: '',
  dest_addr: '',
  content: '',
  encoding: 'GSM7' as 'GSM7' | 'UCS2'
})

// Template state
const templateDialogVisible = ref(false)
const manageDialogVisible = ref(false)
const editDialogVisible = ref(false)
const templates = ref<MessageTemplate[]>([])
const templatesLoading = ref(false)
const savingTemplate = ref(false)
const editingTemplate = ref<MessageTemplate | null>(null)
const templateFormRef = ref<FormInstance>()

const templateForm = ref({
  name: '',
  content: '',
  encoding: 'UCS2' as 'GSM7' | 'UCS2'
})

const templateRules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  content: [{ required: true, message: '请输入模板内容', trigger: 'blur' }]
}

// Deliver methods
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
  deliverForm.value.session_id = receiver.id
  ElMessage.success(`已选择: ${receiver.system_id}`)
}

const handleDeliverSend = async () => {
  if (!deliverFormRef.value) return

  await deliverFormRef.value.validate(async (valid) => {
    if (!valid) return

    sending.value = true
    try {
      await sendApi.sendMessage({
        session_id: deliverForm.value.session_id,
        source_addr: deliverForm.value.source_addr,
        dest_addr: deliverForm.value.dest_addr,
        content: deliverForm.value.content,
        encoding: deliverForm.value.encoding
      })
      ElMessage.success('消息发送成功')
      deliverForm.value.content = ''
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } }
      ElMessage.error(err.response?.data?.error || '发送失败')
    } finally {
      sending.value = false
    }
  })
}

const handleDeliverReset = () => {
  deliverFormRef.value?.resetFields()
}

// Outbound methods
const fetchOutboundSessions = async () => {
  outboundLoading.value = true
  try {
    const { data } = await outboundApi.list()
    outboundSessions.value = data.data || []
  } catch {
    ElMessage.error('获取连接列表失败')
  } finally {
    outboundLoading.value = false
  }
}

const handleConnect = async () => {
  if (!connectFormRef.value) return

  connecting.value = true
  try {
    const { data } = await outboundApi.connect({
      host: connectForm.value.host,
      port: connectForm.value.port,
      system_id: connectForm.value.system_id,
      password: connectForm.value.password,
      bind_type: connectForm.value.bind_type
    })
    ElMessage.success('连接成功')
    fetchOutboundSessions()
    // 自动选中新连接的会话
    if (data.data) {
      selectedOutboundSession.value = data.data
    }
  } catch (error: unknown) {
    const err = error as { response?: { data?: { error?: string } } }
    ElMessage.error(err.response?.data?.error || '连接失败')
  } finally {
    connecting.value = false
  }
}

const handleDisconnect = async (session: OutboundSession) => {
  try {
    await outboundApi.disconnect(session.id)
    ElMessage.success('已断开连接')
    if (selectedOutboundSession.value?.id === session.id) {
      selectedOutboundSession.value = null
    }
    fetchOutboundSessions()
  } catch {
    ElMessage.error('断开连接失败')
  }
}

const selectOutboundSession = (session: OutboundSession) => {
  if (session.status !== 'active') {
    ElMessage.warning('该连接不可用')
    return
  }
  selectedOutboundSession.value = session
  outboundSendForm.value.source_addr = ''
  outboundSendForm.value.dest_addr = ''
  outboundSendForm.value.content = ''
}

const handleOutboundSend = async () => {
  if (!selectedOutboundSession.value) return

  if (!outboundSendForm.value.source_addr || !outboundSendForm.value.dest_addr || !outboundSendForm.value.content) {
    ElMessage.warning('请填写完整的发送信息')
    return
  }

  outboundSending.value = true
  try {
    await outboundApi.sendMessage(selectedOutboundSession.value.id, {
      source_addr: outboundSendForm.value.source_addr,
      dest_addr: outboundSendForm.value.dest_addr,
      content: outboundSendForm.value.content,
      encoding: outboundSendForm.value.encoding
    })
    ElMessage.success('消息发送成功')
    outboundSendForm.value.content = ''
  } catch (error: unknown) {
    const err = error as { response?: { data?: { error?: string } } }
    ElMessage.error(err.response?.data?.error || '发送失败')
  } finally {
    outboundSending.value = false
  }
}

// Template methods
const fetchTemplates = async () => {
  templatesLoading.value = true
  try {
    const { data } = await templateApi.list()
    templates.value = data.data || []
  } catch {
    ElMessage.error('获取模板列表失败')
  } finally {
    templatesLoading.value = false
  }
}

const showTemplateDialog = () => {
  templateDialogVisible.value = true
  fetchTemplates()
}

const selectTemplate = (template: MessageTemplate) => {
  deliverForm.value.content = template.content
  if (template.encoding === 'UCS2' || template.encoding === 'GSM7') {
    deliverForm.value.encoding = template.encoding
  }
  templateDialogVisible.value = false
  ElMessage.success(`已应用模板: ${template.name}`)
}

const showManageDialog = () => {
  templateDialogVisible.value = false
  manageDialogVisible.value = true
  fetchTemplates()
}

const showCreateDialog = () => {
  editingTemplate.value = null
  templateForm.value = {
    name: '',
    content: '',
    encoding: 'UCS2'
  }
  editDialogVisible.value = true
}

const editTemplate = (template: MessageTemplate) => {
  editingTemplate.value = template
  templateForm.value = {
    name: template.name,
    content: template.content,
    encoding: (template.encoding === 'GSM7' || template.encoding === 'UCS2') ? template.encoding : 'UCS2'
  }
  editDialogVisible.value = true
}

const saveTemplate = async () => {
  if (!templateFormRef.value) return

  await templateFormRef.value.validate(async (valid) => {
    if (!valid) return

    savingTemplate.value = true
    try {
      if (editingTemplate.value) {
        await templateApi.update(editingTemplate.value.id, {
          name: templateForm.value.name,
          content: templateForm.value.content,
          encoding: templateForm.value.encoding
        })
        ElMessage.success('模板更新成功')
      } else {
        await templateApi.create({
          name: templateForm.value.name,
          content: templateForm.value.content,
          encoding: templateForm.value.encoding
        })
        ElMessage.success('模板创建成功')
      }
      editDialogVisible.value = false
      fetchTemplates()
    } catch {
      ElMessage.error('保存模板失败')
    } finally {
      savingTemplate.value = false
    }
  })
}

const deleteTemplate = async (template: MessageTemplate) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${template.name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await templateApi.delete(template.id)
    ElMessage.success('模板删除成功')
    fetchTemplates()
  } catch {
    // User cancelled or API error
  }
}

onMounted(() => {
  fetchReceivers()
  fetchOutboundSessions()
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

.send-tabs {
  margin-bottom: 20px;
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

.content-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
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

.outbound-sessions-card {
  margin-bottom: 20px;
}

.outbound-send-card {
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

.template-manage-header {
  margin-bottom: 16px;
}

/* 移动端接收方列表 */
.mobile-receiver-list,
.mobile-outbound-list {
  display: none;
}

.receiver-item,
.outbound-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #EBEEF5;
  cursor: pointer;
}

.receiver-item:last-child,
.outbound-item:last-child {
  border-bottom: none;
}

.receiver-item:active,
.outbound-item:active {
  background-color: #f5f7fa;
}

.receiver-info,
.outbound-info {
  flex: 1;
  min-width: 0;
}

.receiver-name,
.outbound-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.receiver-addr,
.outbound-addr {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.outbound-error {
  font-size: 12px;
  color: #F56C6C;
  margin-top: 2px;
}

.receiver-meta,
.outbound-meta {
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

  .receivers-card,
  .outbound-sessions-card {
    margin-bottom: 16px;
  }

  .desktop-table {
    display: none;
  }

  .mobile-receiver-list,
  .mobile-outbound-list {
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
