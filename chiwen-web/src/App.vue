<template>
  <router-view v-slot="{ Component, route }">
    <!-- 登录页不使用布局 -->
    <component :is="Component" v-if="route.path === '/login'" />

    <!-- 其他页面使用 MainLayout 布局 -->
    <main-layout v-else>
      <component :is="Component" />
    </main-layout>
  </router-view>
</template>

<script setup lang="ts">
import MainLayout from '@/layout/MainLayout.vue'
</script>

<style>
/* 全局淡入淡出动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>