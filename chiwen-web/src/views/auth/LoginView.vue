<!-- src/views/auth/LoginView.vue -->
<template>
  <div class="login-container">
    <a-card class="login-box">
      <!-- 标题区 -->
      <template #title>
        <div class="login-header">
          <h1 class="title">
            <span class="logo-text">螭吻</span>
            <span class="subtitle">运维平台</span>
          </h1>
          <p class="desc">Chiwen Web Terminal System</p>
        </div>
      </template>

      <!-- 登录表单 -->
      <a-form
        ref="loginFormRef"
        :model="form"
        :rules="rules"
        size="large"
        @submit.prevent="handleLogin"
      >
        <a-form-item field="username" :validate-trigger="['change', 'blur']">
          <a-input
            v-model="form.username"
            placeholder="请输入用户名"
            allow-clear
          >
            <template #prefix><UserOutlined /></template>
          </a-input>
        </a-form-item>

        <a-form-item field="password" :validate-trigger="['change', 'blur']">
          <a-input-password
            v-model="form.password"
            placeholder="请输入密码"
            allow-clear
          >
            <template #prefix><LockOutlined /></template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            size="large"
            class="login-btn"
            :loading="loading"
            html-type="submit"
          >
            {{ loading ? '登录中...' : '立即登录' }}
          </a-button>
        </a-form-item>
      </a-form>

      <!-- 页脚 -->
      <div class="footer">
        © 2025 Chiwen WebTTY System. All rights reserved.
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { message } from 'ant-design-vue'  // 正确导入 Ant Design Vue 的 message
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const loginFormRef = ref()

const form = reactive({
  username: 'admin',
  password: 'admin123'
})

const rules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }]
}

const handleLogin = async () => {
  try {
    await loginFormRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const success = await authStore.login(form.username.trim(), form.password)

    if (success) {
      message.success('登录成功！')
      router.push('/dashboard')
    } else {
      message.error('用户名或密码错误')
    }
  } catch (error) {
    message.error('网络异常，请检查后端服务是否启动')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="less">
.login-container {
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-box {
  width: 100%;
  max-width: 420px;
  border-radius: 12px;
  overflow: hidden;
  background: white;  // Antd card 默认白底

  .login-header {
    text-align: center;
    margin-bottom: 30px;

    .title {
      font-size: 32px;
      font-weight: 700;
      color: #001529;
      margin: 0 0 8px 0;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;

      .logo-text {
        color: #1890ff;
        font-size: 42px;
        font-weight: 900;
      }

      .subtitle {
        font-size: 20px;
      }
    }

    .desc {
      color: rgba(0, 0, 0, 0.65);
      margin: 0;
      font-size: 14px;
    }
  }

  .login-btn {
    width: 100%;
    height: 48px;
    font-size: 16px;
    border-radius: 8px;
  }

  .footer {
    text-align: center;
    margin-top: 30px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }
}
</style>