<template>
  <div class="admin-view">
    <el-card class="admin-card" header="注册审批管理">
      <div class="filter-bar">
        <el-button type="primary" @click="fetchPendingApplies" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新列表
        </el-button>
        <div class="stats">
          <el-tag type="info">待审批: {{ pendingCount }}</el-tag>
        </div>
      </div>

      <el-table
        :data="pendingApplies"
        :loading="loading"
        row-key="id"
        size="medium"
        style="width: 100%; margin-top: 20px"
      >
        <el-table-column
          prop="id"
          label="客户端ID"
          width="280"
        >
          <template #default="{ row }">
            <el-tooltip :content="row.id" placement="top">
              <span class="id-text">{{ formatID(row.id) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
          prop="hostname"
          label="主机名"
          width="180"
        />
        <el-table-column
          prop="nonce"
          label="Nonce"
          width="120"
        >
          <template #default="{ row }">
            <el-tag size="small">{{ formatNonce(row.nonce) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="created_at"
          label="申请时间"
          width="180"
        >
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="client_public_key"
          label="公钥"
          width="120"
        >
          <template #default="{ row }">
            <el-tooltip :content="row.client_public_key" placement="top">
              <el-tag type="info" size="small">查看</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="200"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button-group>
              <el-button type="success" size="small" @click="handleApprove(row)" :loading="approvingId === row.id">
                批准
              </el-button>
              <el-button type="danger" size="small" @click="handleReject(row)" :loading="rejectingId === row.id">
                拒绝
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="pendingApplies.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无待审批的申请" />
      </div>
    </el-card>

    <!-- 审批确认对话框 -->
    <el-dialog
      v-model="approveDialogVisible"
      title="批准申请"
      width="500px"
      :close-on-click-modal="false"
    >
      <div v-if="currentApply">
        <p>确定要批准以下客户端的注册申请吗？</p>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="客户端ID">{{ currentApply.id }}</el-descriptions-item>
          <el-descriptions-item label="主机名">{{ currentApply.hostname }}</el-descriptions-item>
          <el-descriptions-item label="申请时间">{{ formatDate(currentApply.created_at) }}</el-descriptions-item>
        </el-descriptions>
        <p style="margin-top: 15px; color: #67c23a;">
          <el-icon><InfoFilled /></el-icon>
          批准后，客户端将获得访问权限并出现在资产列表中。
        </p>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="approveDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmApprove" :loading="approving">
            确定批准
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 拒绝确认对话框 -->
    <el-dialog
      v-model="rejectDialogVisible"
      title="拒绝申请"
      width="500px"
      :close-on-click-modal="false"
    >
      <div v-if="currentApply">
        <p>确定要拒绝以下客户端的注册申请吗？</p>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="客户端ID">{{ currentApply.id }}</el-descriptions-item>
          <el-descriptions-item label="主机名">{{ currentApply.hostname }}</el-descriptions-item>
          <el-descriptions-item label="申请时间">{{ formatDate(currentApply.created_at) }}</el-descriptions-item>
        </el-descriptions>
        <p style="margin-top: 15px; color: #f56c6c;">
          <el-icon><WarningFilled /></el-icon>
          拒绝后，客户端将无法连接，申请记录将被删除。
        </p>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="rejectDialogVisible = false">取消</el-button>
          <el-button type="danger" @click="confirmReject" :loading="rejecting">
            确定拒绝
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { 
  ElCard, 
  ElTable, 
  ElTableColumn, 
  ElTag, 
  ElButton, 
  ElButtonGroup,
  ElDialog,
  ElDescriptions,
  ElDescriptionsItem,
  ElMessage,
  ElMessageBox,
  ElEmpty,
  ElTooltip,
  ElIcon
} from 'element-plus';
import { Refresh, InfoFilled, WarningFilled } from '@element-plus/icons-vue';
import { useAuthStore } from '@/stores/auth';

const authStore = useAuthStore();
const loading = ref(false);
const pendingApplies = ref([]);
const currentApply = ref(null);
const approveDialogVisible = ref(false);
const rejectDialogVisible = ref(false);
const approving = ref(false);
const rejecting = ref(false);
const approvingId = ref('');
const rejectingId = ref('');

const pendingCount = computed(() => pendingApplies.value.length);

const formatID = (id) => {
  if (!id) return '';
  if (id.length > 20) {
    return id.substring(0, 8) + '...' + id.substring(id.length - 8);
  }
  return id;
};

const formatNonce = (nonce) => {
  if (!nonce) return '';
  if (nonce.length > 10) {
    return nonce.substring(0, 6) + '...';
  }
  return nonce;
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  try {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN');
  } catch (error) {
    return dateString;
  }
};

const fetchPendingApplies = async () => {
  loading.value = true;
  try {
    const response = await fetch('/api/v1/register/pending', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    pendingApplies.value = data.applies || [];
    
    if (data.count === 0) {
      ElMessage.info('暂无待审批的申请');
    }
  } catch (error) {
    console.error('获取待审批申请失败:', error);
    ElMessage.error('获取待审批申请失败: ' + error.message);
  } finally {
    loading.value = false;
  }
};

const handleApprove = (apply) => {
  currentApply.value = apply;
  approveDialogVisible.value = true;
};

const handleReject = (apply) => {
  currentApply.value = apply;
  rejectDialogVisible.value = true;
};

const confirmApprove = async () => {
  if (!currentApply.value) return;
  
  approving.value = true;
  approvingId.value = currentApply.value.id;
  
  try {
    const response = await fetch('/api/v1/approve', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ id: currentApply.value.id })
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    
    if (data.status === 'approved') {
      ElMessage.success('申请已批准');
      // 从列表中移除已批准的申请
      pendingApplies.value = pendingApplies.value.filter(
        apply => apply.id !== currentApply.value.id
      );
      approveDialogVisible.value = false;
    } else {
      ElMessage.warning('审批状态异常: ' + data.status);
    }
  } catch (error) {
    console.error('批准申请失败:', error);
    ElMessage.error('批准申请失败: ' + error.message);
  } finally {
    approving.value = false;
    approvingId.value = '';
    currentApply.value = null;
  }
};

const confirmReject = async () => {
  if (!currentApply.value) return;
  
  rejecting.value = true;
  rejectingId.value = currentApply.value.id;
  
  try {
    const response = await fetch('/api/v1/register/reject', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ id: currentApply.value.id })
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    
    if (data.status === 'rejected') {
      ElMessage.success('申请已拒绝');
      // 从列表中移除已拒绝的申请
      pendingApplies.value = pendingApplies.value.filter(
        apply => apply.id !== currentApply.value.id
      );
      rejectDialogVisible.value = false;
    } else {
      ElMessage.warning('拒绝状态异常: ' + data.status);
    }
  } catch (error) {
    console.error('拒绝申请失败:', error);
    ElMessage.error('拒绝申请失败: ' + error.message);
  } finally {
    rejecting.value = false;
    rejectingId.value = '';
    currentApply.value = null;
  }
};

onMounted(() => {
  fetchPendingApplies();
});
</script>

<style scoped>
.admin-view {
  padding: 20px;
}

.admin-card {
  width: 100%;
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.stats {
  display: flex;
  gap: 10px;
}

.id-text {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
  color: #666;
  cursor: default;
}

.empty-state {
  margin: 40px 0;
  text-align: center;
}

:deep(.el-table) {
  width: 100%;
}

:deep(.el-table__header-wrapper th) {
  background-color: #fafafa;
  font-weight: 600;
}
</style>
