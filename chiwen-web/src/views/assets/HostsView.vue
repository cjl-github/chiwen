<template>
  <div class="hosts-view">
    <el-card class="hosts-card" header="主机资产列表">
      <el-table
        :data="assets"
        :loading="loading"
        row-key="id"
        size="medium"
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          label="ID"
          width="80"
        />
        <el-table-column
          prop="hostname"
          label="主机名"
          width="150"
        />
        <el-table-column
          prop="status"
          label="状态"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="created_at"
          label="创建时间"
          width="180"
        />
        <el-table-column
          prop="updated_at"
          label="更新时间"
          width="180"
        />
        <el-table-column
          prop="labels"
          label="标签"
          width="150"
        >
          <template #default="{ row }">
            <span v-if="row.labels">{{ formatJson(row.labels) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="allowed_users"
          label="允许用户"
          width="150"
        >
          <template #default="{ row }">
            <span v-if="row.allowed_users && row.allowed_users.length > 0">{{ row.allowed_users.join(', ') }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="static_info"
          label="静态信息"
          width="150"
        >
          <template #default="{ row }">
            <span v-if="row.static_info">{{ formatJson(row.static_info) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="dynamic_info"
          label="动态信息"
          width="150"
        >
          <template #default="{ row }">
            <span v-if="row.dynamic_info">{{ formatJson(row.dynamic_info) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="is_deleted"
          label="已删除"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="row.is_deleted ? 'danger' : 'success'">
              {{ row.is_deleted ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue';
import { useAssetsStore } from '@/stores/assets';
import { ElCard, ElTable, ElTableColumn, ElTag, ElMessage } from 'element-plus';

const assetsStore = useAssetsStore();
const loading = ref(false);

const assets = computed(() => assetsStore.assets);

const getStatusType = (status) => {
  switch (status) {
    case 'online':
      return 'success';
    case 'offline':
      return 'danger';
    case 'maintenance':
      return 'warning';
    default:
      return 'info';
  }
};

const formatJson = (obj) => {
  try {
    if (typeof obj === 'string') {
      return obj;
    }
    return JSON.stringify(obj, null, 2);
  } catch {
    return String(obj);
  }
};

onMounted(async () => {
  loading.value = true;
  try {
    await assetsStore.fetchAssets();
    if (assetsStore.error) {
      ElMessage.error('加载资产数据失败: ' + assetsStore.error);
    }
  } catch (error) {
    console.error('加载资产数据失败:', error);
    ElMessage.error('加载资产数据失败');
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.hosts-view {
  padding: 20px;
}

.hosts-card {
  width: 100%;
}

:deep(.el-table) {
  width: 100%;
}

:deep(.el-table__header-wrapper th) {
  background-color: #fafafa;
  font-weight: 600;
}
</style>
