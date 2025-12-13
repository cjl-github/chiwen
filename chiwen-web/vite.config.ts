// chiwen-web/vite.config.ts
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    host: '0.0.0.0',  // 允许局域网访问（可选，便于手机/其他设备测试）
    port: 5173,       // 前端端口
    open: true,       // 启动时自动打开浏览器
    proxy: {
      '/api': {
        target: 'http://localhost:8090',  // chiwen-server 后端运行端口（根据实际启动端口修改）
        changeOrigin: true,              // 必需：修改请求头中的 Origin，避免后端 CORS 拒绝
        secure: false,                   // 如果后端不是 HTTPS，可保持 false
        // rewrite: (path) => path.replace(/^\/api/, ''),  // 如果后端路径不带 /api 前缀，可取消注释此行
      },
      // Agent WebSocket 连接（如 TTY 会话、agent 心跳等）通常走 /ws 或类似路径
      '/ws': {
        target: 'http://localhost:8090',
        ws: true,              // 启用 WebSocket 代理
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
