<template>
  <div class="system-page">
    <h1 class="page-title">系统配置</h1>

    <el-card v-loading="loading">
      <template #header>
        <span>系统信息</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="SMPP端口">
          <el-tag>{{ config.smpp_port }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="HTTP端口">
          <el-tag>{{ config.http_port }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据库类型">
          <el-tag type="info">{{ config.db_type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Redis状态">
          <el-tag :type="config.redis_enabled ? 'success' : 'info'">
            {{ config.redis_status }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Security Settings -->
    <el-card class="settings-card">
      <template #header>
        <span>安全设置</span>
      </template>
      <el-form :model="form" label-width="160px" :rules="rules" ref="formRef">
        <el-divider content-position="left">修改密码</el-divider>
        
        <el-form-item label="当前密码" prop="old_password">
          <el-input
            v-model="form.old_password"
            type="password"
            placeholder="输入当前密码"
            show-password
            clearable
          />
        </el-form-item>
        
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="form.new_password"
            type="password"
            placeholder="输入新密码（至少6位）"
            show-password
            clearable
          />
        </el-form-item>
        
        <el-form-item label="确认新密码" prop="confirm_password">
          <el-input
            v-model="form.confirm_password"
            type="password"
            placeholder="再次输入新密码"
            show-password
            clearable
          />
        </el-form-item>

        <el-divider content-position="left">其他设置</el-divider>

        <el-form-item label="JWT过期时间">
          <el-input-number
            v-model="form.jwt_expiry"
            :min="1"
            :max="720"
            :step="1"
          />
          <span class="form-unit">小时</span>
          <div class="form-tip">Token有效期，范围1-720小时</div>
        </el-form-item>

        <el-form-item label="CORS允许来源">
          <el-input
            v-model="form.cors_origins"
            placeholder="多个来源用逗号分隔，如: http://localhost:3000,https://example.com"
            clearable
          />
          <div class="form-tip">允许跨域请求的来源，* 表示允许所有</div>
        </el-form-item>

        <el-form-item label="登录限流">
          <el-input-number
            v-model="form.login_rate_limit"
            :min="1"
            :max="100"
            :step="1"
          />
          <span class="form-unit">次/分钟</span>
          <div class="form-tip">每分钟最大登录尝试次数</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">
            保存配置
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Security Warning -->
    <el-card class="warning-card">
      <template #header>
        <span>安全提示</span>
      </template>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
      >
        <template #title>
          <span>请定期更新管理员密码，确保系统安全</span>
        </template>
        <ul class="warning-list">
          <li>密码修改后立即生效，所有已登录用户需要重新登录</li>
          <li>JWT过期时间修改后，新生成的Token将使用新的过期时间</li>
          <li>CORS设置修改后需要重启服务才能生效</li>
          <li>请勿使用默认密码或弱密码</li>
        </ul>
      </el-alert>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { systemApi } from '@/api'
import type { SystemConfig } from '@/types'

const loading = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const config = ref<SystemConfig>({
  smpp_port: '',
  http_port: '',
  db_type: '',
  redis_enabled: false,
  redis_status: '',
  admin_password: '',
  jwt_expiry: 24,
  cors_origins: '*',
  login_rate_limit: 5
})

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
  jwt_expiry: 24,
  cors_origins: '*',
  login_rate_limit: 5
})

// Password validation rule
const validatePassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (form.new_password && !value) {
    callback(new Error('请输入当前密码'))
  } else {
    callback()
  }
}

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (form.new_password && value !== form.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  old_password: [{ validator: validatePassword, trigger: 'blur' }],
  new_password: [
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirm_password: [{ validator: validateConfirmPassword, trigger: 'blur' }]
}

const fetchConfig = async () => {
  loading.value = true
  try {
    const response = await systemApi.getConfig()
    config.value = response.data
    // Copy to form
    form.jwt_expiry = response.data.jwt_expiry
    form.cors_origins = response.data.cors_origins
    form.login_rate_limit = response.data.login_rate_limit
  } catch {
    ElMessage.error('获取配置失败')
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    // Check if changing password
    if (form.new_password && !form.old_password) {
      ElMessage.error('修改密码需要输入当前密码')
      return
    }

    // Confirm for sensitive changes
    if (form.new_password) {
      try {
        await ElMessageBox.confirm(
          '修改密码后，所有已登录用户需要重新登录。确定要继续吗？',
          '确认修改',
          { type: 'warning' }
        )
      } catch {
        return
      }
    }

    saving.value = true
    try {
      const updateData: Record<string, unknown> = {
        jwt_expiry: form.jwt_expiry,
        cors_origins: form.cors_origins,
        login_rate_limit: form.login_rate_limit
      }

      // Only include password fields if changing password
      if (form.new_password) {
        updateData.old_password = form.old_password
        updateData.new_password = form.new_password
        updateData.confirm_password = form.confirm_password
      }

      await systemApi.updateConfig(updateData)
      ElMessage.success('配置已保存')
      
      // Clear password fields
      form.old_password = ''
      form.new_password = ''
      form.confirm_password = ''
      
      // Refresh config
      await fetchConfig()
    } catch {
      // Error handled by interceptor
    } finally {
      saving.value = false
    }
  })
}

const handleReset = () => {
  form.old_password = ''
  form.new_password = ''
  form.confirm_password = ''
  form.jwt_expiry = config.value.jwt_expiry
  form.cors_origins = config.value.cors_origins
  form.login_rate_limit = config.value.login_rate_limit
  formRef.value?.clearValidate()
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.system-page {
  max-width: 800px;
}

.page-title {
  font-size: 24px;
  margin-bottom: 20px;
  color: #303133;
}

.settings-card {
  margin-top: 20px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.form-unit {
  margin-left: 8px;
  color: #606266;
}

.warning-card {
  margin-top: 20px;
}

.warning-list {
  margin: 8px 0 0 0;
  padding-left: 20px;
  line-height: 1.8;
}

.warning-list li {
  color: #e6a23c;
}

:deep(.el-divider__text) {
  font-weight: 500;
  color: #303133;
}
</style>
