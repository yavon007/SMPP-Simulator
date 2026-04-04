<template>
  <div class="login-page">
    <div class="login-background">
      <div class="login-container">
        <div class="login-header">
          <el-icon class="login-icon"><Cellphone /></el-icon>
          <h1>SMPP Simulator</h1>
          <p>短信协议模拟器</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="login-form"
          @submit.prevent="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="用户名"
              size="large"
              :prefix-icon="User"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              size="large"
              :prefix-icon="Lock"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="login-button"
              @click="handleLogin"
            >
              {{ loading ? '登录中...' : '登 录' }}
            </el-button>
          </el-form-item>

          <el-alert
            v-if="error"
            :title="error"
            type="error"
            :closable="false"
            class="login-error"
          />
        </el-form>

        <div class="login-footer">
          <p>默认用户名: admin</p>
          <p>请使用配置文件或环境变量设置密码</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Cellphone } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: 'admin',
  password: ''
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    error.value = ''

    try {
      await authStore.login(form.username, form.password)
      ElMessage.success('登录成功')

      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } catch (err: any) {
      error.value = err.response?.data?.error || '登录失败，请检查用户名和密码'
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-page {
  min-height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-background {
  width: 100%;
  max-width: 420px;
  padding: 20px;
}

.login-container {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 40px 32px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-icon {
  font-size: 48px;
  color: #409EFF;
  margin-bottom: 16px;
}

.login-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.login-header p {
  font-size: 14px;
  color: #909399;
}

.login-form {
  width: 100%;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 8px;
}

.login-button {
  width: 100%;
  border-radius: 8px;
  font-size: 16px;
  height: 44px;
}

.login-error {
  margin-top: 16px;
  border-radius: 8px;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #EBEEF5;
}

.login-footer p {
  font-size: 12px;
  color: #909399;
  margin: 4px 0;
}

/* 移动端适配 */
@media (max-width: 480px) {
  .login-container {
    padding: 32px 20px;
  }

  .login-icon {
    font-size: 40px;
  }

  .login-header h1 {
    font-size: 20px;
  }
}
</style>
