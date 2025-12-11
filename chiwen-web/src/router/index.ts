// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// 1. 公开路由：不需要登录就能访问
const publicRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue')
  }
]

// 2. 需要登录 + 使用统一后台布局的路由
const privateRoutes = [
  {
    path: '/',
    component: () => import('@/layout/MainLayout.vue'), // ← 所有后台页面都走这个布局
    children: [
      { path: '', redirect: '/dashboard' }, // 根路径跳转到仪表盘
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue')
      },
      {
        path: 'assets',
        name: 'Assets',
        component: () => import('@/views/assets/AssetsList.vue')
      },
      {
        path: 'sessions',
        name: 'Sessions',
        component: () => import('@/views/sessions/SessionList.vue')
      },
      {
        path: 'audit',
        name: 'Audit',
        component: () => import('@/views/audit/AuditList.vue')
      }
      // 以后想加新页面，直接在这里加就行
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [...publicRoutes, ...privateRoutes]
})

// ==================== 全局路由守卫（保持你原来的逻辑）====================
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  authStore.init() // 从 localStorage 读取 token

  // 需要登录的页面（所有 privateRoutes里都有 MainLayout 的都算需要登录）
  const requiresAuth = to.matched.some(record => record.path !== '/login')

  if (requiresAuth && !authStore.token) {
    // 未登录 → 去登录页
    next('/login')
  } else if (to.path === '/login' && authStore.token) {
    // 已登录却访问登录页 → 跳转仪表盘
    next('/dashboard')
  } else {
    next()
  }
})

export default router