<template>
  <div class="admin-view">
    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h2>管理系统</h2>
          <el-tag type="warning">正在开发中</el-tag>
        </div>
      </template>
      
      <div class="coming-soon">
        <el-empty description="功能正在开发中，敬请期待">
          <template #image>
            <div class="empty-image">
              <el-icon size="100" color="#e6a23c"><Setting /></el-icon>
            </div>
          </template>
        </el-empty>
        
        <div class="features-list">
          <h3>计划中的功能：</h3>
          <el-timeline>
            <el-timeline-item
              v-for="(feature, index) in plannedFeatures"
              :key="index"
              :timestamp="feature.timestamp"
              :type="feature.type"
              placement="top"
            >
              <el-card shadow="hover">
                <h4>{{ feature.title }}</h4>
                <p>{{ feature.description }}</p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { ElCard, ElTag, ElEmpty, ElTimeline, ElTimelineItem, ElIcon } from 'element-plus';
import { Setting } from '@element-plus/icons-vue';

const plannedFeatures = ref([
  {
    title: '用户管理',
    description: '管理系统用户、角色和权限分配',
    timestamp: '2025年Q1',
    type: 'primary'
  },
  {
    title: '系统配置',
    description: '全局系统参数配置和个性化设置',
    timestamp: '2025年Q1',
    type: 'primary'
  },
  {
    title: '数据备份',
    description: '系统数据备份与恢复功能',
    timestamp: '2025年Q2',
    type: 'success'
  },
  {
    title: '审计日志',
    description: '完整的操作审计日志和报表',
    timestamp: '2025年Q2',
    type: 'success'
  },
  {
    title: '系统监控',
    description: '实时系统性能监控和告警',
    timestamp: '2025年Q3',
    type: 'info'
  }
]);
</script>

<style scoped>
.admin-view {
  padding: 20px;
}

.admin-card {
  width: 100%;
  min-height: 500px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.coming-soon {
  text-align: center;
  padding: 40px 0;
}

.empty-image {
  margin-bottom: 20px;
}

.features-list {
  margin-top: 40px;
  max-width: 800px;
  margin-left: auto;
  margin-right: auto;
}

.features-list h3 {
  text-align: left;
  margin-bottom: 20px;
  color: #333;
}

:deep(.el-timeline-item__timestamp) {
  color: #666;
  font-size: 14px;
}

:deep(.el-timeline-item__content h4) {
  margin: 0 0 8px 0;
  color: #333;
}

:deep(.el-timeline-item__content p) {
  margin: 0;
  color: #666;
  line-height: 1.5;
}
</style>
