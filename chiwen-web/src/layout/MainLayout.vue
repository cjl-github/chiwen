<template>
  <div id="app-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <!-- Logo 区域：精确 200 × 40px -->
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
      <!-- Header 高度改为 40px -->
      <header class="header">
        <h1 class="title">{{ currentTitle }}</h1>
        <a-button type="primary" danger size="small" @click="logout">退出登录</a-button>
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

const expandedMenus = ref<Record<string, boolean>>({ assets: true })

const toggleMenu = (key: string) => {
  expandedMenus.value[key] = !expandedMenus.value[key]
}

const isAssetsActive = computed(() => route.path.startsWith('/assets/'))

const currentTitle = computed(() => {
  const map: Record<string, string>  = {
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
  return map[route.path] || '螭吻平台'
})

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
  grid-template-columns: 200px 1fr;
  overflow: hidden;
}

/* 侧边栏整体 */
.sidebar {
  background: #001529;
  color: #fff;
  display: flex;
  flex-direction: column;
}

/* Logo 区域：200 × 40px */
.logo {
  width: 200px;
  height: 40px;
  background: rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 12px;
  box-sizing: border-box;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}
.logo h2 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: 0.5px;
}

/* 菜单区域 */
.menu {
  flex: 1;
  overflow-y: auto;
}
.menu ul {
  list-style: none;
  padding: 8px 0;
  margin: 0;
}

/* 一级菜单项 */
.menu-item,
.menu-title {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  color: rgba(255, 255, 255, 0.65);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s;
  cursor: pointer;
}
.menu-item:hover,
.menu-title:hover {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}
.menu-item.active,
.menu-title.active,
.menu-title.expanded {
  background: #00c4b6;
  color: white;
}
.icon {
  font-size: 17px;
  width: 26px;
  text-align: center;
  margin-right: 12px;
  flex-shrink: 0;
}
.arrow {
  margin-left: auto;
  font-size: 12px;
  transition: transform 0.3s;
}
.expanded .arrow {
  transform: rotate(180deg);
}

/* 二级菜单 */
.submenu {
  background: #000c17;
}
.submenu a {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  transition: all 0.3s;
}
.submenu a .icon {
  margin-right: 12px;
  width: 26px;
  text-align: center;
}
.submenu a:hover,
.submenu a.active {
  background: #00c4b6;
  color: white;
}

/* 动画 */
.slide-enter-active,
.slide-leave-active { transition: all 0.3s ease; }
.slide-enter-from,
.slide-leave-to { opacity: 0; transform: translateY(-8px); }

/* 右侧区域 */
.main { 
  display: flex; 
  flex-direction: column; 
  background: #f0f2f5; 
}

/* Header 高度改为 40px */
.header { 
  height: 40px;                    /* 关键：从 64px → 40px */
  background: #fff; 
  padding: 0 24px;                 /* 左右内边距略减小 */
  display: flex; 
  align-items: center; 
  justify-content: space-between; 
  box-shadow: 0 1px 4px rgba(0,21,41,.08); 
  border-bottom: 1px solid #e8e8e8;
}

.title { 
  margin: 0; 
  font-size: 16px;                 /* 文字稍小，与 40px 高度匹配 */
  font-weight: 600; 
  color: #000; 
}

/* 退出按钮适配 40px 高度 */
.header :deep(.ant-btn) {
  height: 28px;
  line-height: 26px;
  font-size: 13px;
  padding: 0 12px;
}

.content { 
  flex: 1; 
  padding: 24px 32px; 
  overflow-y: auto; 
  background: #f0f2f5; 
}
</style>