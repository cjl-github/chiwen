<template>
  <div class="admin-view">
    <h1>管理系统</h1>
    <p>这里是管理系统的控制面板，例如用户管理、权限设置。</p>
    <!-- 示例：按钮组 -->
    <div>
      <button @click="handleUserManage">用户管理</button>
      <button @click="handleRoleEdit">角色编辑</button>
    </div>
  </div>
</template>

<script setup lang="ts">
const handleUserManage = () => {
  alert('打开用户管理面板')
}
const handleRoleEdit = () => {
  alert('打开角色编辑面板')
}
</script>

<style scoped>
.admin-view {
  padding: 20px;
}
button {
  margin: 10px;
  padding: 10px 20px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
}
</style>