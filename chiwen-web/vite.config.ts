// vite.config.ts   ← 完整可直接覆盖使用的最终版
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

  // ==================== 关键修改：开发服务器代理 ====================
  server: {
    host: '0.0.0.0',           // 允许局域网访问（可选）
    port: 5173,                // 前端端口，可改成 3000、8080 随便你
    open: true,
    // 所有 /api 开头的请求全部自动转发到你的后端
    proxy: {
      '/api': {
        target: 'http://localhost:8090',   // ←←← 改成你实际后端 IP 和端口
        changeOrigin: true,
        secure: false,
        // 可选：如果你前端代码里写的是 /api/v1/...，这一行可以省略
        // rewrite: (path) => path.replace(/^\/api/, '/api')
      },
      // 如果你以后还有 WebSocket 长连接，也可以顺手加一条
      '/ws': {
        target: 'http://localhost:8090',
        ws: true,
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
