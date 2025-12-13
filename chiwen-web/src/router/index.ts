import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth'; // 从stores/auth.ts导入

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
          path: '',
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
        {
          path: 'assets/assets-list',
          name: 'assets-list',
          component: () => import('@/views/assets/AssetsList.vue'),
          meta: { title: '资产列表' }
        },

        // 其他一级页面
        {
          path: 'sessions',
          name: 'sessions',
          component: () => import('@/views/sessions/SessionList.vue'),
          meta: { title: '会话列表' }
        },
        {
          path: 'pipelines',
          name: 'pipelines',
          component: () => import('@/views/pipeline/PipelineList.vue'),
          meta: { title: '流水线' }
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
    },

    // 直接访问资产路由的重定向（为了兼容旧链接）
    {
      path: '/assets/hosts',
      redirect: '/dashboard/assets/hosts'
    },
    {
      path: '/assets/network',
      redirect: '/dashboard/assets/network'
    },
    {
      path: '/assets/database',
      redirect: '/dashboard/assets/database'
    },
    {
      path: '/assets/cert',
      redirect: '/dashboard/assets/cert'
    },
    {
      path: '/pipelines',
      redirect: '/dashboard/pipelines'
    },
    {
      path: '/monitoring',
      redirect: '/dashboard/monitoring'
    },
    {
      path: '/logs',
      redirect: '/dashboard/logs'
    },
    {
      path: '/audit',
      redirect: '/dashboard/audit'
    },
    {
      path: '/admin',
      redirect: '/dashboard/admin'
    }
  ]
});

// 添加全局守卫，确保登陆后访问
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    next('/login');
  } else {
    next();
  }
});

export default router;
