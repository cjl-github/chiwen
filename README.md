# chiwen

## 整体架构
前端: Vue 3 + TypeScript + Vite + Pinia + Element Plus
后端: Go + Gin + MySQL + JWT
通信: HTTP REST API + WebSocket（用于TTY功能）

登录流程
前端
    用户访问 `http://localhost:5173`
    用户输入用户名和密码（默认：admin/admin123）
    点击"立即登录"按钮或按Enter键
    前端验证表单
    发送登录请求 `/api/v1/login` → Vite代理 → `http://localhost:8090/api/v1/login`
后端
    1 接收请求: Gin框架接收POST请求到 `/api/v1/login`
      参数验证: 绑定JSON到 `LoginRequest` 结构体
      CORS处理: 通过CORS中间件允许跨域请求
    2 用户验证
      查询用户
      密码验证（bcrypt哈希比较）
      生成24小时有效的JWT Token
      响应返回
        JWT Token
        用户信息（排除密码哈希）
前端
    1 前端保存到Pinia/LocalStorage
      跳转到Dashboard页面

### 前端错误处理
- 网络错误：显示"网络异常，请检查后端服务是否启动"
- 认证错误：显示"用户名或密码错误"
- Token过期：自动跳转到登录页面
### 后端错误处理
- 参数错误：返回400状态码
- 用户不存在/禁用：返回401状态码
- 密码错误：返回401状态码
- 服务器错误：返回500状态码

### 持久化存储
- Token保存到 `localStorage`
- 页面刷新时从 `localStorage` 恢复状态
- 登出时清除 `localStorage`


技术栈：前端用 Vue 3 + TypeScript + Vite + Pinia + Element Plus

仪表盘
资产管理
流水线系统
监控系统
日志系统
审计系统
管理系统

：用户数据 资产数据

    主机：名称 ip 账号 配置信息 系统 状态 备注 操作（编辑/删除）
    网络设备：名称 地址 账号 配置信息 系统 状态 备注 操作
    数据库：名称 地址 账号 配置信息 系统 备注 操作
    证书：
    用户管理
        用户列表（平台用户）
        用户组（平台用户组）
        账号列表（对端linux/数据库等账号用作与server端通信建立链接用）
    登陆日志
    系统设置：
        安全设置：MFA认证 IP校验 登陆IP绑定
        LDAP设置：LDAP服务地址 端口 账号 密码
        推送服务：
    审计管理
    资产授权
        名称 用户 用户组 资产 账号 有效时间 操作







我的项目：目录结构
chiwen-web
├── src
│   ├── App.vue
│   ├── layout
│   │   └── MainLayout.vue
│   ├── main.ts
│   ├── router
│   │   └── index.ts
│   ├── stores
│   │   ├── auth.ts
│   │   └── counter.ts
│   └── views
│       ├── assets
│       │   └── AssetsList.vue
│       ├── audit
│       │   └── AuditList.vue
│       ├── auth
│       │   └── LoginView.vue
│       ├── DashboardView.vue
│       └── sessions
│           └── SessionList.vue
具体文件如下
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/main.ts
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/App.vue
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/layout/MainLayout.vue
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/router/index.ts
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/stores/auth.ts
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/stores/counter.ts
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/views/DashboardView.vue
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/views/sessions/SessionList.vue
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/views/auth/LoginView.vue
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/views/audit/AuditList.vue
https://github.com/cjl-github/chiwen/blob/main/chiwen-web/src/views/assets/AssetsList.vue

根据我的 项目 先教我 一步 一步实现 这个侧边栏的优化
仪表盘
资产管理
流水线系统
监控系统
日志系统
审计系统
管理系统
  