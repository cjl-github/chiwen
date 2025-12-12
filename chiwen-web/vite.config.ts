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
    port: 5173,       // 前端端口，可根据需要修改
    open: true,       // 自动打开浏览器
    proxy: {
      '/api': {
        target: 'http://localhost:8090',  // 后端地址（如果后端在其他 IP/端口，改这里）
        changeOrigin: true,               // 修改 origin 头，避免后端 CORS 校验
        secure: false,                    // 如果后端是 https，改成 true
        // 无需 rewrite，因为前后端路径一致 (/api/v1/login)
      },
      // 如果有 WebSocket（如 tty/ws），加这一条
      '/ws': {
        target: 'http://localhost:8090',
        ws: true,
        changeOrigin: true,
        secure: false,
      },
    },
  },
})