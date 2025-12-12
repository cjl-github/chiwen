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
      component: () => import('@/layout/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        // 仪表盘
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { title: '仪表盘' }
        },

        // ========== 资产管理及其子路由 ==========
        {
          path: 'assets/hosts',
          name: 'assets-hosts',
          component: () => import('@/views/assets/HostsView.vue'),
          meta: { title: '主机设备' }
        },
        {
          path: 'assets/network',
          name: 'assets-network',
          component: () => import('@/views/assets/NetworkView.vue'),
          meta: { title: '网络设备' }
        },
        {
          path: 'assets/database',
          name: 'assets-database',
          component: () => import('@/views/assets/DatabaseView.vue'),
          meta: { title: '数据库' }
        },
        {
          path: 'assets/cert',
          name: 'assets-cert',
          component: () => import('@/views/assets/CertView.vue'),
          meta: { title: '证书' }
        },

        // 其他一级页面
        {
          path: 'pipelines',
          name: 'pipelines',
          component: () => import('@/views/sessions/SessionList.vue'),
          meta: { title: '流水线系统' }
        },
        {
          path: 'monitoring',
          name: 'monitoring',
          component: () => import('@/views/monitoring/MonitoringView.vue'),
          meta: { title: '监控系统' }
        },
        {
          path: 'logs',
          name: 'logs',
          component: () => import('@/views/logs/LogsView.vue'),
          meta: { title: '日志系统' }
        },
        {
          path: 'audit',
          name: 'audit',
          component: () => import('@/views/audit/AuditList.vue'),
          meta: { title: '审计系统' }
        },
        {
          path: 'admin',
          name: 'admin',
          component: () => import('@/views/admin/AdminView.vue'),
          meta: { title: '管理系统' }
        }
      ]
    }
  ]
})

// 登录守卫（你原来的逻辑，完全保留）
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router