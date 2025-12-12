<template>
  <!-- 仪表盘页面内容全部清空 -->
  <div class="empty-dashboard">
    <!-- 你可以留一句提示，也可以完全什么都不写 -->
    <!-- <h3>仪表盘已关闭</h3>
    <p>如需恢复原内容，请查看原始 DashboardView.vue 备份。</p> -->

    <!-- 如果你希望彻底空白，连上面两行文字都删掉即可 -->
  </div>
</template>

<script setup lang="ts">
// 这里原来可能有大量请求接口、ECharts 初始化、定时器等代码
// 全部都不需要了，直接保持空即可
// 如果你之前备份过原文件，现在可以放心删光 script 内容
</script>

<style scoped>
/* 保持背景和原来的内容区一致，避免页面突变 */
.empty-dashboard {
  width: 100%;
  height: 100%;
  background: #f7f9fc; /* 与 MainLayout.vue 中 .content 的背景色一致 */
  /* 可选：居中显示提示文字 */
  /* display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #999;
  font-size: 18px; */
}

/* 如果你发现还有黄色边框（可能是全局卡片样式），强制清除 */
:deep(.el-card),
:deep(.el-card__body) {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
}
</style>