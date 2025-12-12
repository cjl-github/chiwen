<template>
  <div id="app-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="logo">
        <h2>螭吻平台</h2>
      </div>

      <nav class="menu">
        <ul>
          <!-- 仪表盘 -->
          <li>
            <router-link to="/dashboard" class="menu-item" active-class="active">
              <HomeOutlined class="icon" />
              <span>仪表盘</span>
            </router-link>
          </li>

          <!-- 资产管理（可展开） -->
          <li class="menu-group">
            <div
              class="menu-title"
              :class="{ 'expanded': expandedMenus.assets, 'active': isAssetsActive }"
              @click="toggleMenu('assets')"
            >
              <DatabaseOutlined class="icon" />
              <span>资产管理</span>
              <DownOutlined class="arrow" />
            </div>

            <Transition name="slide">
              <ul v-show="expandedMenus.assets" class="submenu">
                <li>
                  <router-link to="/assets/hosts" active-class="active">
                    <DesktopOutlined class="icon" />
                    <span>主机设备</span>
                  </router-link>
                </li>
                <li>
                  <router-link to="/assets/network" active-class="active">
                    <WifiOutlined class="icon" />
                    <span>网络设备</span>
                  </router-link>
                </li>
                <li>
                  <router-link to="/assets/database" active-class="active">
                    <DatabaseOutlined class="icon" />
                    <span>数据库</span>
                  </router-link>
                </li>
                <li>
                  <router-link to="/assets/cert" active-class="active">
                    <SafetyCertificateOutlined class="icon" />
                    <span>证书</span>
                  </router-link>
                </li>
              </ul>
            </Transition>
          </li>

          <!-- 其他一级菜单 -->
          <li>
            <router-link to="/pipelines" class="menu-item" active-class="active">
              <ForkOutlined class="icon" />
              <span>流水线系统</span>
            </router-link>
          </li>
          <li>
            <router-link to="/monitoring" class="menu-item" active-class="active">
              <FundOutlined class="icon" />
              <span>监控系统</span>
            </router-link>
          </li>
          <li>
            <router-link to="/logs" class="menu-item" active-class="active">
              <FileTextOutlined class="icon" />
              <span>日志系统</span>
            </router-link>
          </li>
          <li>
            <router-link to="/audit" class="menu-item" active-class="active">
              <AuditOutlined class="icon" />
              <span>审计系统</span>
            </router-link>
          </li>
          <li>
            <router-link to="/admin" class="menu-item" active-class="active">
              <SettingOutlined class="icon" />
              <span>管理系统</span>
            </router-link>
          </li>
        </ul>
      </nav>
    </aside>

    <!-- 右侧内容区 -->
    <div class="main">
      <header class="header">
        <h1 class="title">{{ currentTitle }}</h1>
        <a-button type="primary" danger @click="logout">退出登录</a-button>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  HomeOutlined,
  DatabaseOutlined,
  DesktopOutlined,
  WifiOutlined,
  SafetyCertificateOutlined,
  ForkOutlined,
  FundOutlined,
  FileTextOutlined,
  AuditOutlined,
  SettingOutlined,
  DownOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const expandedMenus = ref({ assets: true })

const toggleMenu = (key: string) => {
  expandedMenus.value[key] = !expandedMenus.value[key]
}

// 当前路由是否在资产管理模块下
const isAssetsActive = computed(() => route.path.startsWith('/assets/'))

// 动态标题
const currentTitle = computed(() => {
  const map: Record<string, string> = {
    '/dashboard': '仪表盘',
    '/assets/hosts': '主机设备',
    '/assets/network': '网络设备',
    '/assets/database': '数据库',
    '/assets/cert': '证书',
    '/pipelines': '流水线系统',
    '/monitoring': '监控系统',
    '/logs': '日志系统',
    '/audit': '审计系统',
    '/admin': '管理系统',
  }
  return map[route.path] || '迟文系统'
})

// 进入资产子页面自动展开
watch(
  () => route.path,
  (path) => {
    if (path.startsWith('/assets/')) expandedMenus.value.assets = true
  },
  { immediate: true }
)

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>




<style scoped>
#app-layout {
  height: 100vh;
  display: grid;
  grid-template-columns: 240px 1fr;
  overflow: hidden;
}

/* 侧边栏 */
.sidebar {
  background: #001529;
  color: #fff;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 64px;
  background: rgba(255,255,255,0.05);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 600;
}

/* 菜单容器 */
.menu ul {
  list-style: none;
  padding: 8px 0;
  margin: 0;
  flex: 1;
}

/* 一级菜单项 + 父菜单标题（完全统一） */
.menu-item,
.menu-title {
  display: flex;
  align-items: center;
  padding: 13px 24px;                    /* 上下左右内边距一致 */
  color: rgba(255,255,255,0.65);
  font-size: 14.5px;                     /* 一二级文字大小完全一致 */
  font-weight: 500;
  transition: all 0.3s;
  cursor: pointer;
}
.menu-item:hover,
.menu-title:hover {
  background: rgba(255,255,255,0.08);
  color: white;
}

/* 高亮状态：青绿底白字 */
.menu-item.active,
.menu-title.active,
.menu-title.expanded {
  background: #00c4b6;
  color: white;
}

/* 图标固定宽度，确保文字完全对齐 */
.icon {
  font-size: 18px;
  width: 28px;                 /* 关键！固定宽度，所有图标对齐 */
  text-align: center;
  margin-right: 16px;
  flex-shrink: 0;
}

/* 箭头 */
.arrow {
  margin-left: auto;
  font-size: 12px;
  transition: transform 0.3s;
}
.expanded .arrow {
  transform: rotate(180deg);
}

/* 二级菜单：完全和一级菜单对齐 + 同款文字大小 */
.submenu {
  background: #000c17;
  padding: 0;
}
.submenu a {
  display: flex;
  align-items: center;
  padding: 13px 24px;                    /* 和一级菜单完全一样的 padding */
  color: rgba(255,255,255,0.7);
  font-size: 14.5px;                     /* 文字大小完全一致！ */
  transition: all 0.3s;
}
.submenu a .icon {
  margin-right: 16px;
  width: 28px;
  text-align: center;
}
.submenu a:hover,
.submenu a.active {
  background: #00c4b6;
  color: white;
}

/* 展开动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 右侧保持不变 */
.main { display: flex; flex-direction: column; background: #f0f2f5; }
.header { height: 64px; background: #fff; padding: 0 32px; display: flex; align-items: center; justify-content: space-between; box-shadow: 0 1px 4px rgba(0,21,41,.08); }
.title { margin: 0; font-size: 20px; font-weight: 600; color: #000; }
.content { flex: 1; padding: 24px 32px; overflow-y: auto; background: #f0f2f5; }
</style>