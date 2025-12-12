// chiwen-web/src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>('')           
  const user = ref<any>(null)             
  const userRole = ref<string>('')        

  const isLoggedIn = computed(() => !!token.value)

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
    }
    return permissions[userRole.value]?.includes(key) ?? key === 'dashboard'
  }

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      const res = await fetch('/api/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      if (!res.ok) {
        console.error('登录响应失败:', res.status, await res.text());
        return false;
      }

      const data = await res.json()

      if (data.token) {
        token.value = data.token
        user.value = data.user || null
        // 修复：使用后端 is_admin 字段判断角色（原代码用 role，但后端无此字段）
        userRole.value = data.user?.is_admin ? 'admin' : 'user'
        localStorage.setItem('token', token.value)
        return true
      }

      return false
    } catch (error) {
      console.error('登录失败:', error)
      return false
    }
  }

  const logout = () => {
    token.value = ''
    user.value = null
    userRole.value = ''
    localStorage.removeItem('token')
  }

  const initAuth = () => {
    const savedToken = localStorage.getItem('token')
    if (savedToken) {
      token.value = savedToken
      // TODO: 可添加 /api/v1/me 接口刷新用户信息
    }
  }

  return {
    token,
    user,
    userRole,
    isLoggedIn,
    isAuthenticated: isLoggedIn,
    login,
    logout,
    initAuth,
    hasPermission,
  }
})