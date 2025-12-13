import { defineStore } from 'pinia';
import axios from 'axios';
import { useAuthStore } from './auth'; // 假设 auth store 有 token 和 isLoggedIn

// 定义资产类型接口（匹配后端API返回的数据结构）
export interface Asset {
  ID: string;
  Hostname: string;
  Status: 'online' | 'offline' | 'maintenance';
  CreatedAt: string;
  UpdatedAt: string;
  Labels?: string | null;
  AllowedUsers?: string | null;
  StaticInfo?: string | null;
  DynamicInfo?: string | null;
  ClientPubKey?: string;
  IsDeleted?: boolean;
}

export const useAssetsStore = defineStore('assets', {
  state: () => ({
    assets: [] as Asset[], // 显式类型，避免 never[] 报错
    loading: false,
    error: null as string | null,
  }),
  getters: {
    onlineAssets: (state) => state.assets.filter((a) => a.Status === 'online'),
  },
  actions: {
    async fetchAssets() {
      const authStore = useAuthStore();
      if (!authStore.token) {
        this.error = '未登录';
        return;
      }

      this.loading = true;
      this.error = null;

      try {
        // 使用正确的 API 路径
        const response = await axios.get<Asset[]>('/api/v1/assets/list', {
          headers: { Authorization: `Bearer ${authStore.token}` },
        });

        this.assets = response.data;
      } catch (err: any) {
        console.error('获取资产失败:', err);
        this.error = err.response?.data?.message || '请求失败';
      } finally {
        this.loading = false;
      }
    },

    async deleteAsset(assetId: string) {
      const authStore = useAuthStore();
      if (!authStore.token) {
        this.error = '未登录';
        return false;
      }

      try {
        await axios.delete(`/api/v1/assets/${assetId}`, {
          headers: { Authorization: `Bearer ${authStore.token}` },
        });
        
        // 从本地状态中移除已删除的资产
        this.assets = this.assets.filter(asset => asset.ID !== assetId);
        return true;
      } catch (err: any) {
        console.error('删除资产失败:', err);
        this.error = err.response?.data?.message || '删除失败';
        return false;
      }
    },

    async updateAssetLabels(assetId: string, labels: Record<string, any>) {
      const authStore = useAuthStore();
      if (!authStore.token) {
        this.error = '未登录';
        return false;
      }

      try {
        await axios.put(`/api/v1/assets/${assetId}/labels`, 
          { labels },
          {
            headers: { 
              Authorization: `Bearer ${authStore.token}`,
              'Content-Type': 'application/json'
            },
          }
        );
        
        // 更新本地状态
        const assetIndex = this.assets.findIndex(asset => asset.ID === assetId);
        if (assetIndex !== -1 && this.assets[assetIndex]) {
          this.assets[assetIndex].Labels = JSON.stringify(labels);
        }
        
        return true;
      } catch (err: any) {
        console.error('更新资产备注失败:', err);
        this.error = err.response?.data?.message || '更新失败';
        return false;
      }
    },
  },
});
