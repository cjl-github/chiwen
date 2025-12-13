import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { useAuthStore } from './stores/auth'  // 新增：导入 auth store

const app = createApp(App)

app.use(createPinia())

// 新增：恢复 auth 状态（必须在 use(Pinia) 后）
const authStore = useAuthStore()
authStore.initAuth()

app.use(router)
app.use(ElementPlus)

// 全局注册Element Plus图标组件
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')
