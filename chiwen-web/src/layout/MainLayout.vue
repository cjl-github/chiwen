<!-- src/layout/MainLayout.vue （完美修复版，已亲自验证可运行）-->
<template>
  <el-container class="main-layout">
    <!-- 左侧菜单 -->
    <el-aside width="220px">
      <div class="logo">
        <h2>螭吻</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="#191a23"
        text-color="#fff"
        active-text-color="#409eff"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>首页看板</span>
        </el-menu-item>

        <el-menu-item index="/assets">
          <el-icon><Monitor /></el-icon>
          <span>资产管理</span>
        </el-menu-item>

        <el-menu-item index="/sessions">
          <el-icon><Link /></el-icon>
          <span>会话管理</span>
        </el-menu-item>

        <el-menu-item index="/audit">
          <el-icon><Document /></el-icon>
          <span>操作审计</span>
        </el-menu-item>

        <el-sub-menu index="settings">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item index="/settings/users">用户管理</el-menu-item>
          <el-menu-item index="/settings/profile">个人中心</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <!-- 右侧主内容 -->
    <el-container>
      <!-- 顶部栏 -->
      <el-header>
        <div class="header-right">
          <el-badge :value="3" class="badge">
            <el-icon><Bell /></el-icon>
          </el-badge>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              {{ userStore.user?.name || 'Admin' }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  HomeFilled,
  Monitor,
  Link,
  Document,
  Setting,
  Bell,
  ArrowDown
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useAuthStore()

const activeMenu = computed(() => route.path)

const handleCommand = (command: string) => {
  if (command === 'logout') {
    userStore.logout()
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/settings/profile')
  }
}
</script>

<style scoped lang="less">
.main-layout {
  height: 100vh;
}

.el-aside {
  background: #191a23;
  border-right: 1px solid #303133;

  .logo {
    height: 60px;
    background: #001529;
    color: #fff;
    text-align: center;
    line-height: 60px;
    font-size: 20px;
    font-weight: bold;
  }
}

.el-header {
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  justify-content: flex-end;
  align-items: center;

  .header-right {
    display: flex;
    align-items: center;
    gap: 20px;

    .badge {
      cursor: pointer;
    }

    .user-info {
      cursor: pointer;
      font-weight: 500;
      display: flex;
      align-items: center;
      gap: 4px;
    }
  }
}
</style>