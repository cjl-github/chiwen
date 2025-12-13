<template>
  <div class="q-pa-md">
    <q-table
      title="主机资产列表"
      :rows="assets"
      :columns="columns"
      row-key="id"
      :loading="loading"
      :pagination="{ rowsPerPage: 10 }"
    >
      <template v-slot:loading>
        <q-inner-loading showing color="primary" />
      </template>
    </q-table>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { useAssetsStore } from '@/stores/assets'; // 导入store

const assetsStore = useAssetsStore();
const assets = ref([]);
const loading = ref(false);

const columns = [
  { name: 'id', label: 'ID', field: 'id', sortable: true },
  { name: 'hostname', label: '主机名', field: 'hostname', sortable: true },
  { name: 'status', label: '状态', field: 'status', sortable: true },
  { name: 'created_at', label: '创建时间', field: 'created_at', sortable: true },
  { name: 'updated_at', label: '更新时间', field: 'updated_at', sortable: true },
  { name: 'labels', label: '标签', field: 'labels', format: val => val ? JSON.stringify(val) : '' },
  { name: 'allowed_users', label: '允许用户', field: 'allowed_users', format: val => val ? val.join(', ') : '' },
  { name: 'static_info', label: '静态信息', field: 'static_info', format: val => val ? JSON.stringify(val) : '' },
  { name: 'dynamic_info', label: '动态信息', field: 'dynamic_info', format: val => val ? JSON.stringify(val) : '' },
  { name: 'is_deleted', label: '已删除', field: 'is_deleted', format: val => val ? '是' : '否' }
  // 可以根据需要添加更多列，如client_public_key（但公钥可能敏感，不建议显示完整），agent_secret_key（秘密，不显示）
];

onMounted(async () => {
  loading.value = true;
  try {
    await assetsStore.fetchAssets(); // 调用store action拉取数据
    assets.value = assetsStore.assets; // 从store绑定数据
  } catch (error) {
    console.error('加载资产数据失败:', error);
    // 可选：使用Quasar notify显示错误
    // this.$q.notify({ type: 'negative', message: '加载失败' });
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
/* 可选样式，例如调整表格宽度或颜色 */
.q-table {
  width: 100%;
}
</style>