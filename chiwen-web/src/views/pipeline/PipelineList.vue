<template>
  <div class="pipeline-page">
    <h1>管道列表</h1>
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
.pipeline-page {
  padding: 40px;
  text-align: center;
  font-size: 18px;
}
</style>
