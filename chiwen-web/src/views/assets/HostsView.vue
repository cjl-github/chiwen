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
          prop="Hostname"
          label="名称"
          width="150"
        />
        <el-table-column
          label="IP"
          width="120"
        >
          <template #default="{ row }">
            <span>{{ extractIP(row.StaticInfo) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="账号"
          width="100"
        >
          <template #default="{ row }">
            <span>{{ extractAccount(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="配置信息"
          width="150"
        >
          <template #default="{ row }">
            <span>{{ extractConfig(row.StaticInfo) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="系统"
          width="120"
        >
          <template #default="{ row }">
            <span>{{ extractOS(row.StaticInfo) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="Status"
          label="状态"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.Status)">
              {{ getStatusText(row.Status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="备注"
          width="150"
        >
          <template #default="{ row }">
            <span>{{ extractRemark(row.Labels) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="150"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑备注"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="备注">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="4"
            placeholder="请输入备注信息"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveEdit" :loading="saving">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, reactive } from 'vue';
import { useAssetsStore } from '@/stores/assets';
import { 
  ElCard, 
  ElTable, 
  ElTableColumn, 
  ElTag, 
  ElMessage, 
  ElMessageBox,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElButton
} from 'element-plus';

const assetsStore = useAssetsStore();
const loading = ref(false);
const editDialogVisible = ref(false);
const saving = ref(false);
const currentAssetId = ref('');

const assets = computed(() => assetsStore.assets);

// 编辑表单
const editForm = reactive({
  remark: ''
});

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

const getStatusText = (status) => {
  switch (status) {
    case 'online':
      return '在线';
    case 'offline':
      return '离线';
    case 'maintenance':
      return '维护中';
    default:
      return status;
  }
};

const extractIP = (staticInfo) => {
  if (!staticInfo) return '-';
  
  try {
    const info = typeof staticInfo === 'string' ? JSON.parse(staticInfo) : staticInfo;
    
    // 优先使用internal_ips字段（从CollectStaticInfo收集）
    if (info.internal_ips && Array.isArray(info.internal_ips) && info.internal_ips.length > 0) {
      // 返回第一个内网IP
      return info.internal_ips[0];
    }
    
    // 尝试从网络接口中提取IP（兼容旧数据）
    if (info.network && info.network.interfaces) {
      const interfaces = info.network.interfaces;
      // 优先查找eth0或en0等主要接口
      const primaryInterface = interfaces.find((iface) => 
        iface.name && (iface.name.includes('eth') || iface.name.includes('en') || iface.name.includes('wlan'))
      );
      if (primaryInterface && primaryInterface.addresses && primaryInterface.addresses.length > 0) {
        // 优先返回IPv4地址
        const ipv4 = primaryInterface.addresses.find(addr => addr.family === 'IPv4');
        if (ipv4) return ipv4.address;
        // 如果没有IPv4，返回第一个地址
        return primaryInterface.addresses[0].address;
      }
    }
    
    // 如果网络接口中没有找到，尝试其他字段
    if (info.ip) return info.ip;
    if (info.ip_address) return info.ip_address;
    if (info.host_ip) return info.host_ip;
    
    return '-';
  } catch (error) {
    console.error('解析static_info失败:', error);
    return '-';
  }
};

const extractAccount = (row) => {
  if (!row) return '-';
  
  // 尝试从AllowedUsers中提取第一个用户
  if (row.AllowedUsers && typeof row.AllowedUsers === 'string') {
    try {
      const users = JSON.parse(row.AllowedUsers);
      if (Array.isArray(users) && users.length > 0) {
        return users[0];
      }
    } catch (error) {
      console.error('解析AllowedUsers失败:', error);
    }
  }
  
  // 默认返回root
  return 'root';
};

const extractConfig = (staticInfo) => {
  if (!staticInfo) return '-';
  
  try {
    const info = typeof staticInfo === 'string' ? JSON.parse(staticInfo) : staticInfo;
    
    // 优先使用config字段（从CollectStaticInfo收集，格式为"cpu:Xc mem:Yg"）
    if (info.config) {
      return info.config;
    }
    
    // 兼容旧数据：如果没有config字段，尝试从其他字段提取
    const parts = [];
    
    // 提取CPU信息
    if (info.cpu) {
      if (info.cpu.model) {
        parts.push(info.cpu.model.split('@')[0].trim()); // 只取型号部分
      } else if (info.cpu.cores) {
        parts.push(`${info.cpu.cores}核`);
      }
    }
    
    // 提取内存信息
    if (info.memory && info.memory.total) {
      const totalMB = info.memory.total / (1024 * 1024);
      if (totalMB >= 1024) {
        parts.push(`${Math.round(totalMB / 1024)}GB`);
      } else {
        parts.push(`${Math.round(totalMB)}MB`);
      }
    }
    
    // 提取磁盘信息
    if (info.disk && info.disk.total) {
      const totalGB = info.disk.total / (1024 * 1024 * 1024);
      parts.push(`${Math.round(totalGB)}GB`);
    }
    
    return parts.length > 0 ? parts.join('/') : '-';
  } catch (error) {
    console.error('解析static_info配置失败:', error);
    return '-';
  }
};

const extractOS = (staticInfo) => {
  if (!staticInfo) return '-';
  
  try {
    const info = typeof staticInfo === 'string' ? JSON.parse(staticInfo) : staticInfo;
    
    // 优先使用os_detail字段（从CollectStaticInfo收集）
    if (info.os_detail) {
      let osDetail = info.os_detail;
      // 简化OS名称
      if (osDetail.includes('Ubuntu')) return 'Ubuntu';
      if (osDetail.includes('CentOS')) return 'CentOS';
      if (osDetail.includes('Debian')) return 'Debian';
      if (osDetail.includes('Windows')) return 'Windows';
      if (osDetail.includes('macOS')) return 'macOS';
      if (osDetail.includes('Linux')) return 'Linux';
      return osDetail;
    }
    
    // 兼容旧数据
    if (info.os) {
      if (info.os.name) {
        let osName = info.os.name;
        // 简化OS名称
        if (osName.includes('Ubuntu')) return 'Ubuntu';
        if (osName.includes('CentOS')) return 'CentOS';
        if (osName.includes('Debian')) return 'Debian';
        if (osName.includes('Windows')) return 'Windows';
        if (osName.includes('macOS')) return 'macOS';
        if (osName.includes('Linux')) return 'Linux';
        return osName;
      }
    }
    
    return '-';
  } catch (error) {
    console.error('解析static_info OS失败:', error);
    return '-';
  }
};

const extractRemark = (labels) => {
  if (!labels) return '-';
  
  try {
    const labelObj = typeof labels === 'string' ? JSON.parse(labels) : labels;
    
    if (typeof labelObj === 'object' && labelObj !== null) {
      // 检查是否是空对象或默认值
      const keys = Object.keys(labelObj);
      if (keys.length === 0) {
        return '-'; // 空对象不显示
      }
      
      // 检查是否是默认的String: {}格式
      if (keys.length === 1 && keys[0] === 'String' && labelObj[String] === '{}') {
        return '-'; // 默认格式不显示
      }
      
      // 尝试提取备注字段
      if (labelObj.remark && labelObj.remark.trim() !== '') return labelObj.remark;
      if (labelObj.description && labelObj.description.trim() !== '') return labelObj.description;
      if (labelObj.note && labelObj.note.trim() !== '') return labelObj.note;
      
      // 如果没有特定字段，检查是否有实际内容
      // 过滤掉空值或默认值
      const nonEmptyKeys = keys.filter(key => {
        const value = labelObj[key];
        return value !== null && value !== undefined && value !== '' && value !== '{}';
      });
      
      if (nonEmptyKeys.length > 0) {
        const firstKey = nonEmptyKeys[0];
        const value = labelObj[firstKey];
        return `${firstKey}: ${value}`;
      }
    }
    
    return '-';
  } catch (error) {
    console.error('解析labels失败:', error);
    return '-';
  }
};

const handleEdit = (row) => {
  currentAssetId.value = row.ID;
  
  // 解析现有的labels，提取remark
  let existingRemark = '';
  if (row.Labels) {
    try {
      const labels = typeof row.Labels === 'string' ? JSON.parse(row.Labels) : row.Labels;
      if (labels && typeof labels === 'object') {
        existingRemark = labels.remark || labels.description || labels.note || '';
      }
    } catch (error) {
      console.error('解析labels失败:', error);
    }
  }
  
  editForm.remark = existingRemark;
  editDialogVisible.value = true;
};

const saveEdit = async () => {
  if (!currentAssetId.value) return;
  
  saving.value = true;
  try {
    // 创建labels对象，包含remark字段
    const labels = {
      remark: editForm.remark.trim(),
      updatedAt: new Date().toISOString()
    };
    
    const success = await assetsStore.updateAssetLabels(currentAssetId.value, labels);
    if (success) {
      ElMessage.success('备注更新成功');
      editDialogVisible.value = false;
    } else {
      ElMessage.error('更新失败: ' + (assetsStore.error || '未知错误'));
    }
  } catch (error) {
    console.error('保存备注失败:', error);
    ElMessage.error('保存失败');
  } finally {
    saving.value = false;
  }
};

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除主机 "${row.Hostname}" 吗？此操作不可撤销。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    );
    
    const success = await assetsStore.deleteAsset(row.ID);
    if (success) {
      ElMessage.success('删除成功');
    } else {
      ElMessage.error('删除失败: ' + (assetsStore.error || '未知错误'));
    }
  } catch (error) {
    // 用户取消删除
    console.log('用户取消删除');
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
