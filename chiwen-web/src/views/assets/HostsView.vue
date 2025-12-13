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
          prop="hostname"
          label="名称"
          width="150"
        />
        <el-table-column
          label="IP"
          width="120"
        >
          <template #default="{ row }">
            <span>{{ extractIP(row.static_info) }}</span>
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
            <span>{{ extractConfig(row.static_info) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="系统"
          width="120"
        >
          <template #default="{ row }">
            <span>{{ extractOS(row.static_info) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          width="100"
        >
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="备注"
          width="150"
        >
          <template #default="{ row }">
            <span>{{ extractRemark(row.labels) }}</span>
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
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue';
import { useAssetsStore } from '@/stores/assets';
import { ElCard, ElTable, ElTableColumn, ElTag, ElMessage, ElMessageBox } from 'element-plus';

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
    
    // 尝试从网络接口中提取IP
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
  
  // 尝试从allowed_users中提取第一个用户
  if (row.allowed_users && Array.isArray(row.allowed_users) && row.allowed_users.length > 0) {
    return row.allowed_users[0];
  }
  
  // 如果allowed_users是字符串，尝试解析
  if (row.allowed_users && typeof row.allowed_users === 'string') {
    try {
      const users = JSON.parse(row.allowed_users);
      if (Array.isArray(users) && users.length > 0) {
        return users[0];
      }
    } catch (error) {
      console.error('解析allowed_users失败:', error);
    }
  }
  
  // 默认返回root
  return 'root';
};

const extractConfig = (staticInfo) => {
  if (!staticInfo) return '-';
  
  try {
    const info = typeof staticInfo === 'string' ? JSON.parse(staticInfo) : staticInfo;
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
      // 尝试提取备注字段
      if (labelObj.remark) return labelObj.remark;
      if (labelObj.description) return labelObj.description;
      if (labelObj.note) return labelObj.note;
      
      // 如果没有特定字段，返回第一个键值对
      const keys = Object.keys(labelObj);
      if (keys.length > 0) {
        const firstKey = keys[0];
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
  ElMessageBox.alert('编辑功能尚未实现', '提示', {
    confirmButtonText: '确定',
    callback: () => {
      console.log('编辑资产:', row);
      // TODO: 实现编辑逻辑
    },
  });
};

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除主机 "${row.hostname}" 吗？此操作不可撤销。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(() => {
      ElMessage.success('删除成功（演示功能）');
      console.log('删除资产:', row);
      // TODO: 实现删除API调用
      // 在实际应用中，这里应该调用删除API，然后重新加载数据
    })
    .catch(() => {
      // 用户取消
    });
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
