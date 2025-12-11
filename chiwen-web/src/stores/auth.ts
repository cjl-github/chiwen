// src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  // ==================== 状态 ====================
  const token = ref<string>('')           // JWT token
  const user = ref<any>(null)             // 后端返回的完整用户信息
  const userRole = ref<string>('')        // 角色标识：admin / user / guest 等

  // ==================== 计算属性 ====================
  // 当前是否已登录
  const isLoggedIn = computed(() => !!token.value)

  //  // 权限判断（侧边栏最常用）
  const hasPermission = (key: string): boolean => {
    const permissions: Record<string, string[]> = {
      admin: [
        'dashboard',
        'assets',
        'pipeline',
        'monitor',
        'logs',
        'audit',
        'management',
      ],
      user: ['dashboard', 'assets', 'audit'],
      // 以后可以继续加 guest / auditor 等角色
    }

    // 如果没找到角色，默认只给 dashboard
    return permissions[userRole.value]?.includes(key) ?? key === 'dashboard'
  }

  // ==================== 方法 ====================
  /** 登录（你原来的 fetch 逻辑） */
  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      const res = await fetch('/api/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      const data = await res.json()

      if (res.ok && data.token) {
        token.value = data.token
        user.value = data.user || null

        // 重点：从后端返回的用户信息里取角色
        // 常见字段：data.user.role / data.role / data.user.role_name
        userRole.value = data.user?.role || data.role || 'user'

        // 持久化
        localStorage.setItem('token', token.value)
        return true
      }

      return false
    } catch (error) {
      console.error('登录失败:', error)
      return false
    }
  }

  /** 退出登录 */
  const logout = () => {
    token.value = ''
    user.value = null
    userRole.value = ''
    localStorage.removeItem('token')
  }

  /** 项目启动时自动恢复 token（必须在 main.ts 中调用） */
  const initAuth = () => {
    const savedToken = localStorage.getItem('token')
    if (savedToken) {
      token.value = savedToken
      // 可选：调用 /api/v1/me 接口重新拉取用户信息和角色
      // 这里先留空，实际项目建议加上
    }
  }

  // ==================== 返回 ====================
  return {
    // 状态
  token,
  user,
  userRole,
  isLoggedIn,
  isAuthenticated: isLoggedIn,   // 加上这一行！以后 authStore.isAuthenticated 也能用
  login,
  logout,
  initAuth,
  hasPermission,
  }
})