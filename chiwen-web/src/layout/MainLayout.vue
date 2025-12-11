<template>
  <div id="main-layout">
    <!-- 顶部栏 -->
    <header class="header">
      <h1>迟文系统</h1>
      <div class="top-right">
        <span>{{ authStore.user?.name || '管理员' }}</span>
        <button @click="logout" class="logout-btn">退出登录</button>
      </div>
    </header>

    <div class="main-container">
      <!-- 侧边栏 -->
      <aside class="sidebar" :class="{ collapsed: isCollapsed }">
        <nav class="menu">
          <ul>
            <li><router-link to="/dashboard"   active-class="active">仪表盘</router-link></li>
            <li><router-link to="/assets"      active-class="active">资产管理</router-link></li>
            <li><router-link to="/pipelines"   active-class="active">流水线系统</router-link></li>
            <li><router-link to="/monitoring"  active-class="active">监控系统</router-link></li>
            <li><router-link to="/logs"          active-class="active">日志系统</router-link></li>
            <li><router-link to="/audit"       active-class="active">审计系统</router-link></li>
            <li><router-link to="/admin"       active-class="active">管理系统</router-link></li>
          </ul>
        </nav>
      </aside>

      <!-- 主内容区 -->
      <main class="content" :class="{ 'no-sidebar': isCollapsed }">
        <router-view />
      </main>

      <!-- 移动端折叠按钮 -->
      <button class="mobile-toggle" @click="isCollapsed = !isCollapsed">
        {{ isCollapsed ? '>>' : '<<' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isCollapsed = ref(false)

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
#main-layout { height: 100vh; display: flex; flex-direction: column; font-family: Arial, sans-serif; }

.header {
  height: 60px;
  background: #fff;
  border-bottom: 1px solid #e1e4e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  z-index: 10;
}

.top-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.logout-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 6px 16px;
  border-radius: 4px;
  cursor: pointer;
}

.main-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 240px;
  background: #2c3e50;
  transition: width 0.3s;
  overflow: hidden;
}

.sidebar.collapsed {
  width: 0;
}

.menu ul {
  list-style: none;
  padding: 20px 0;
  margin: 0;
}

.menu a {
  display: block;
  padding: 14px 24px;
  color: #ecf0f1;
  text-decoration: none;
  transition: background 0.2s;
}

.menu a:hover,
.menu a.active {
  background: #34495e;
  color: #fff;
}

.content {
  flex: 1;
  padding: 24px;
  background: #f7f9fc;
  overflow-y: auto;
  transition: margin-left 0.3s;
  margin-left: 0;
}

.mobile-toggle {
  position: fixed;
  left: 0;
  top: 50%;
  background: #2c3e50;
  color: white;
  border: none;
  width: 32px;
  height: 60px;
  border-radius: 0 6px 6px 0;
  cursor: pointer;
  z-index: 100;
  display: none;
}

@media (max-width: 768px) {
  .sidebar { width: 0; }
  .content { margin-left: 0 !important; }
  .mobile-toggle { display: block; }
}
</style>