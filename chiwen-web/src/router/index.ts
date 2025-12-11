import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/dashboard' },

    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue')
    },

    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { requiresAuth: true }
    },

    {
      path: '/assets',
      name: 'assets',
      component: () => import('@/views/assets/AssetsList.vue'),
      meta: { requiresAuth: true }
    },

    {
      path: '/pipelines',
      name: 'pipelines',
      component: () => import('@/views/sessions/SessionList.vue'),
      meta: { requiresAuth: true }
    },

    {
      path: '/monitoring',
      name: 'monitoring',
      component: () => import('@/views/monitoring/MonitoringView.vue'),
      meta: { requiresAuth: true }
    },

    {
      path: '/logs',
      name: 'logs',
      component: () => import('@/views/logs/LogsView.vue'),
      meta: { requiresAuth: true }
    },

    {
      path: '/audit',
      name: 'audit',
      component: () => import('@/views/audit/AuditList.vue'),
      meta: { requiresAuth: true }
    },

    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/admin/AdminView.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

// 关键修改：使用你真实的状态名 isLoggedIn
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router