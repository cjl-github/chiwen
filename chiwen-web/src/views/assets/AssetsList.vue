<!-- src/views/assets/AssetsList.vue -->
<template>
  <div class="assets-page">
    <h1>资产列表</h1>
    <p>正在开发中……</p>
    <el-button type="primary" @click="goBack">返回 Dashboard</el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
const goBack = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.assets-page {
  padding: 40px;
  text-align: center;
  font-size: 18px;
}
</style>