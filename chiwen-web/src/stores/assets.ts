import { defineStore } from 'pinia';
import axios from 'axios';
import { useAuthStore } from './auth'; // 假设 auth store 有 token 和 isLoggedIn

// 定义资产类型接口（匹配 assets 表结构）
export interface Asset {
  id: string;
  hostname: string;
  status: 'online' | 'offline' | 'maintenance';
  created_at: string;
  updated_at: string;
  labels?: Record<string, any> | null;
  allowed_users?: string[] | null;
  static_info?: Record<string, any> | null;
  dynamic_info?: Record<string, any> | null;
  client_public_key?: string;
  agent_secret_key?: string;
  is_deleted?: number;
}

export const useAssetsStore = defineStore('assets', {
  state: () => ({
    assets: [] as Asset[], // 显式类型，避免 never[] 报错
    loading: false,
    error: null as string | null,
  }),
  getters: {
    onlineAssets: (state) => state.assets.filter((a) => a.status === 'online'),
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
        // 常见后端 API 路径猜测：尝试这些之一（根据控制台 Network 检查实际路径）
        // 如果 404，改为 '/api/v1/assets' 或查看后端日志确认
        const response = await axios.get<Asset[]>('/api/assets/list', {
          headers: { Authorization: `Bearer ${authStore.token}` },
        });

        // 如果后端返回 { code: 0, data: [...] } 结构，改为 this.assets = response.data.data;
        this.assets = response.data;
      } catch (err: any) {
        console.error('获取资产失败:', err);
        this.error = err.response?.data?.message || '请求失败';
      } finally {
        this.loading = false;
      }
    },
  },
});