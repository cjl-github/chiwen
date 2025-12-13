<!-- src/views/assets/HostsView.vue -->
<template>
  <div class="hosts-view">
    <el-row :gutter="20">
      <el-col :span="6">
        <!-- 块1: 树形结构 -->
        <el-tree
          :data="treeData"
          :props="defaultProps"
          @node-click="handleNodeClick"
        />
      </el-col>
      <el-col :span="18">
        <!-- 块2: 表格 -->
        <el-input
          v-model="searchQuery"
          placeholder="搜索名称或 IP"
          class="search-input"
          @input="filterAssets"
        />
        <el-table :data="filteredAssets" style="width: 100%">
          <el-table-column prop="hostname" label="名称" />
          <el-table-column label="IP">
            <template #default="{ row }">{{ row.static_info?.ip || 'N/A' }}</template>
          </el-table-column>
          <el-table-column label="配置">
            <template #default="{ row }">{{ row.static_info?.cpu || 'N/A' }} / {{ row.static_info?.memory || 'N/A' }}</template>
          </el-table-column>
          <el-table-column label="系统">
            <template #default="{ row }">{{ row.static_info?.os || 'Unknown' }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" />
          <el-table-column label="备注">
            <template #default="{ row }">{{ row.labels?.remark || '无' }}</template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-col>
    </el-row>

    <!-- 编辑对话框 (示例，使用 el-dialog) -->
    <el-dialog v-model="editDialogVisible" title="编辑主机">
      <el-form :model="editForm">
        <el-form-item label="名称"><el-input v-model="editForm.hostname" /></el-form-item>
        <!-- 其他字段... -->
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useAssetsStore } from '@/stores/assets';
import { ElMessageBox, ElRow, ElCol, ElTree, ElInput, ElTable, ElTableColumn, ElButton, ElDialog, ElForm, ElFormItem } from 'element-plus';

const assetsStore = useAssetsStore();
const searchQuery = ref(assetsStore.searchQuery);
const editDialogVisible = ref(false);
const editForm = ref({ id: '', hostname: '', /* 其他字段 */ });

const treeData = computed(() => assetsStore.getTreeData());
const filteredAssets = computed(() => assetsStore.filteredAssets);
const defaultProps = { children: 'children', label: 'label' };

onMounted(() => {
  assetsStore.fetchAssets();
});

const filterAssets = () => {
  assetsStore.searchQuery = searchQuery.value;
  assetsStore.filterAssets();
};

const handleNodeClick = (data: any) => {
  if (data.label.includes('Linux')) {
    assetsStore.selectedCategory = 'linux';
  } else if (data.label.includes('Windows')) {
    assetsStore.selectedCategory = 'windows';
  } else {
    assetsStore.selectedCategory = 'all';
  }
  assetsStore.filterAssets();
  // 如果是叶子节点，可以进一步处理，如选中具体主机
};

const handleEdit = (row: any) => {
  editForm.value = { ...row };
  editDialogVisible.value = true;
};

const submitEdit = () => {
  assetsStore.editAsset(editForm.value.id, editForm.value);
  editDialogVisible.value = false;
};

const handleDelete = (row: any) => {
  ElMessageBox.confirm('确认删除此主机?', '警告', { type: 'warning' })
    .then(() => assetsStore.deleteAsset(row.id))
    .catch(() => {});
};
</script>

<style scoped>
.hosts-view {
  padding: 20px;
}
.search-input {
  margin-bottom: 20px;
  width: 300px;
}
</style>
