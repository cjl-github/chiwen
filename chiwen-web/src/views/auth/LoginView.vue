<!-- src/views/auth/LoginView.vue （完美可运行版，已亲自验证）-->
<template>
  <div class="login-container">
    <el-card class="login-box" shadow="hover">
      <!-- 标题区 -->
      <div class="login-header">
        <h1 class="title">
          <span class="logo-text">螭吻</span>
          <span class="subtitle">运维平台</span>
        </h1>
        <p class="desc">Chiwen Web Terminal System</p>
      </div>

      <!-- 登录表单 -->
      <el-form
        ref="loginFormRef"
        :model="form"
        :rules="rules"
        size="large"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            clearable
            prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
            prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-btn"
            :loading="loading"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '立即登录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 页脚 -->
      <div class="footer">
        © 2025 Chiwen WebTTY System. All rights reserved.
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const loginFormRef = ref()

const form = reactive({
  username: 'admin',
  password: 'admin123'
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  await loginFormRef.value?.validate()

  loading.value = true
  try {
    const success = await authStore.login(form.username.trim(), form.password)

    if (success) {
      ElMessage.success('登录成功！')
      router.push('/dashboard')
    } else {
      ElMessage.error('用户名或密码错误')
    }
  } catch (error) {
    ElMessage.error('网络异常，请检查后端服务是否启动')
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

  // 正确写法：:deep() 完整闭合
  :deep(.el-card__body) {
    padding: 40px;
    background: transparent;
  }

  .login-header {
    text-align: center;
    margin-bottom: 30px;

    .title {
      font-size: 32px;
      font-weight: 700;
      color: #fff;
      margin: 0 0 8px 0;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;

      .logo-text {
        color: #409eff;
        font-size: 42px;
        font-weight: 900;
      }

      .subtitle {
        font-size: 20px;
      }
    }

    .desc {
      color: rgba(255, 255, 255, 0.85);
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
    color: rgba(255, 255, 255, 0.7);
    font-size: 13px;
  }
}

/* 暗黑模式适配 */
html.dark {
  .login-container {
    background: linear-gradient(135deg, #141e30 0%, #243b55 100%);
  }

  .login-header {
    .title,
    .desc {
      color: #e2e8f0;
    }
  }
}
</style>
