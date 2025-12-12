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
    console.log('开始登录请求:', username);
    console.log('请求URL:', '/api/v1/login');
    console.log('请求体:', { username, password });
    
    try {
      const res = await fetch('/api/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      console.log('登录响应状态:', res.status, res.statusText);
      console.log('响应头:', Object.fromEntries(res.headers.entries()));
      
      if (!res.ok) {
        const errorText = await res.text();
        console.error('登录响应失败:', res.status, errorText);
        return false;
      }

      const data = await res.json()
      console.log('登录响应数据:', data);

      if (data.token) {
        token.value = data.token
        user.value = data.user || null
        // 修复：使用后端 is_admin 字段判断角色（原代码用 role，但后端无此字段）
        userRole.value = data.user?.is_admin ? 'admin' : 'user'
        localStorage.setItem('token', token.value)
        console.log('登录成功，token已保存');
        console.log('用户角色:', userRole.value);
        return true
      }

      console.log('登录响应中没有token');
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
