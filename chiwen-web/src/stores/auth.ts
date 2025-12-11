// src/stores/auth.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const user = ref<any>(null)

  const login = async (username: string, password: string) => {
    try {
      const res = await fetch('http://localhost:8090/api/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      })

      const data = await res.json()

      if (res.ok && data.code === 0) {
        token.value = data.data.token
        user.value = data.data.user
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
    localStorage.removeItem('token')
  }

  const init = () => {
    const saved = localStorage.getItem('token')
    if (saved) token.value = saved
  }

  return { token, user, login, logout, init }
})