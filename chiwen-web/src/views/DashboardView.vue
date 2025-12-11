<!-- src/views/DashboardView.vue -->
<template>
  <div class="dashboard">
    <h1>欢迎回来，{{ userStore.user?.name || '管理员' }}！</h1>
    
    <el-row :gutter="20" class="stats">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <el-icon :size="40" color="#409EFF"><Server /></el-icon>
            <div>
              <div class="number">28</div>
              <div class="label">总资产数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <el-icon :size="40" color="#67C23A"><CircleCheck /></el-icon>
            <div>
              <div class="number">26</div>
              <div class="label">在线主机</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <!-- 更多统计卡片 -->
    </el-row>

    <div class="quick-actions">
      <el-button type="primary" size="large" @click="$router.push('/assets')">
        管理资产
      </el-button>
      <el-button size="large" @click="openTerminal">
        打开终端
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
const userStore = useAuthStore()

const openTerminal = () => {
  window.open('/terminal', '_blank')
}
</script>

<style scoped lang="less">
.dashboard {
  padding: 20px;
}

.stats {
  margin: 30px 0;
  
  .stat-item {
    display: flex;
    align-items: center;
    gap: 16px;
    
    .number {
      font-size: 32px;
      font-weight: bold;
    }
    .label {
      color: #909399;
    }
  }
}

.quick-actions {
  text-align: center;
  margin-top: 50px;
  
  .el-button + .el-button {
    margin-left: 20px;
  }
}
</style>