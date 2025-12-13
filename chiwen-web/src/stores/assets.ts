// src/stores/assets.ts
import { defineStore } from 'pinia';

export const useAssetsStore = defineStore('assets', {
  state: () => ({
    assets: [] as any[],
    filteredAssets: [] as any[],
    searchQuery: '',
    selectedCategory: 'all',
  }),
  actions: {
    async fetchAssets() {
      try {
        // 模拟数据，因为后端API尚未实现
        this.assets = this.generateMockAssets();
        this.filterAssets();
      } catch (error) {
        console.error('Failed to fetch assets:', error);
      }
    },
    filterAssets() {
      let filtered = this.assets;
      if (this.searchQuery) {
        filtered = filtered.filter(asset =>
          asset.hostname.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
          // 假设 IP 从 static_info 或 dynamic_info 获取
          (asset.static_info?.ip && asset.static_info.ip.toLowerCase().includes(this.searchQuery.toLowerCase()))
        );
      }
      if (this.selectedCategory !== 'all') {
        filtered = filtered.filter(asset => {
          const os = asset.static_info?.os?.toLowerCase() || '';
          return this.selectedCategory === 'linux' ? os.includes('linux') : os.includes('windows');
        });
      }
      this.filteredAssets = filtered;
    },
    getTreeData() {
      const linuxCount = this.assets.filter(a => a.static_info?.os?.toLowerCase().includes('linux')).length;
      const windowsCount = this.assets.filter(a => a.static_info?.os?.toLowerCase().includes('windows')).length;
      return [
        {
          label: '所有设备',
          children: [
            {
              label: `Linux (${linuxCount})`,
              children: this.assets
                .filter(a => a.static_info?.os?.toLowerCase().includes('linux'))
                .map(a => ({ label: a.hostname, id: a.id })),
            },
            {
              label: `Windows (${windowsCount})`,
              children: this.assets
                .filter(a => a.static_info?.os?.toLowerCase().includes('windows'))
                .map(a => ({ label: a.hostname, id: a.id })),
            },
          ],
        },
      ];
    },
    async editAsset(id: string, data: any) {
      try {
        // 模拟编辑
        const index = this.assets.findIndex(asset => asset.id === id);
        if (index !== -1) {
          this.assets[index] = { ...this.assets[index], ...data };
          this.filterAssets();
        }
      } catch (error) {
        console.error('Failed to edit asset:', error);
      }
    },
    async deleteAsset(id: string) {
      try {
        // 模拟删除
        this.assets = this.assets.filter(asset => asset.id !== id);
        this.filterAssets();
      } catch (error) {
        console.error('Failed to delete asset:', error);
      }
    },
    // 生成模拟资产数据
    generateMockAssets() {
      return [
        {
          id: '1',
          hostname: 'web-server-01',
          status: 'online',
          static_info: {
            ip: '192.168.1.100',
            cpu: '4核',
            memory: '8GB',
            os: 'Ubuntu 20.04'
          },
          dynamic_info: {
            cpu_usage: 45.2,
            memory_usage: 67.8,
            disk_usage: 32.1,
            last_check_in: new Date().toISOString()
          },
          labels: {
            remark: '生产Web服务器',
            group: 'web',
            env: 'production'
          }
        },
        {
          id: '2',
          hostname: 'db-server-01',
          status: 'online',
          static_info: {
            ip: '192.168.1.101',
            cpu: '8核',
            memory: '16GB',
            os: 'CentOS 7'
          },
          dynamic_info: {
            cpu_usage: 23.5,
            memory_usage: 45.6,
            disk_usage: 78.9,
            last_check_in: new Date().toISOString()
          },
          labels: {
            remark: '主数据库',
            group: 'database',
            env: 'production'
          }
        },
        {
          id: '3',
          hostname: 'dev-server-01',
          status: 'offline',
          static_info: {
            ip: '192.168.1.102',
            cpu: '2核',
            memory: '4GB',
            os: 'Windows Server 2019'
          },
          dynamic_info: {
            cpu_usage: 0,
            memory_usage: 0,
            disk_usage: 0,
            last_check_in: new Date(Date.now() - 86400000).toISOString() // 1天前
          },
          labels: {
            remark: '开发测试服务器',
            group: 'dev',
            env: 'development'
          }
        },
        {
          id: '4',
          hostname: 'monitor-01',
          status: 'online',
          static_info: {
            ip: '192.168.1.103',
            cpu: '4核',
            memory: '8GB',
            os: 'Debian 11'
          },
          dynamic_info: {
            cpu_usage: 12.3,
            memory_usage: 34.5,
            disk_usage: 56.7,
            last_check_in: new Date().toISOString()
          },
          labels: {
            remark: '监控服务器',
            group: 'monitoring',
            env: 'production'
          }
        },
        {
          id: '5',
          hostname: 'backup-01',
          status: 'maintenance',
          static_info: {
            ip: '192.168.1.104',
            cpu: '4核',
            memory: '8GB',
            os: 'Ubuntu 22.04'
          },
          dynamic_info: {
            cpu_usage: 8.9,
            memory_usage: 21.4,
            disk_usage: 89.2,
            last_check_in: new Date().toISOString()
          },
          labels: {
            remark: '备份服务器',
            group: 'backup',
            env: 'production'
          }
        }
      ];
    }
  },
});
