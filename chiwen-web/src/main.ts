import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'          // Antd 4.x 推荐的样式文件

// 新增这两行：全局注册所有图标（以后随便用，不用每次 import）
import * as Icons from '@ant-design/icons-vue'
import { useAuthStore } from './stores/auth'  // 新增：导入 auth store

const app = createApp(App)

app.use(createPinia())

// 新增：恢复 auth 状态（必须在 use(Pinia) 后）
const authStore = useAuthStore()
authStore.initAuth()

app.use(router)
app.use(Antd)

// 全局注册图标组件
Object.keys(Icons).forEach(key => {
  app.component(key, (Icons as any)[key])
})

app.mount('#app')