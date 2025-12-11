<template>
  <div class="monitoring-view">
    <h1>监控系统</h1>
    <p>这里是监控系统的仪表盘内容，例如实时图表、警报列表等。</p>
    <!-- 示例：添加一个简单的表格 -->
    <table>
      <thead>
        <tr>
          <th>指标</th>
          <th>状态</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>CPU 使用率</td>
          <td>正常</td>
        </tr>
        <tr>
          <td>内存占用</td>
          <td>警告</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
// 无需逻辑，纯展示
</script>

<style scoped>
.monitoring-view {
  padding: 20px;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th, td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}
</style>