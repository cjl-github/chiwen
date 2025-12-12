# 技术栈：前端用 Vue 3 + TypeScript + Vite + Pinia + Element Plus

# 安装 nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"                    # 这句加载 nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # 这句加载自动补全（可选）
source ~/.zshrc

# 查看可以安装的所有 Node 版本
nvm ls-remote

# 安装最新 LTS（强烈推荐大多数人用这个）
nvm install --lts

# 查看本地已经安装的版本
nvm ls
v24.11.1

2023 年后不推荐安装❌
npm install -g @vue/cli 
vue --version
@vue/cli 5.0.9

2025年后推荐 不需要安装 底层使用create-vue
npm init vue@latest   标准命令
npm create vue@latest 别名/快捷方式

npm init vue@latest 
Need to install the following packages:
create-vue@3.18.3
Ok to proceed? (y) y
> npx
> "create-vue"

┌  Vue.js - The Progressive JavaScript Framework
│
◇  请输入项目名称：
│  chiwen-web
│
◇  请选择要包含的功能： (↑/↓ 切换，空格选择，a 全选，回车确认)
│  TypeScript, Router（单页面应用开发）, Pinia（状态管理）, ESLint（错误预防）, Prettier（代码格式化）
│
◇  选择要包含的试验特性： (↑/↓ 切换，空格选择，a 全选，回车确认)
│  none
│
◇  跳过所有示例代码，创建一个空白的 Vue 项目？
│  Yes

正在初始化项目 /Users/levi/Downloads/project/web/chiwen-web...
│
└  项目初始化完成，可执行以下命令：

   cd chiwen-web
   npm install
   npm run format
   npm run dev

| 可选：使用以下命令在项目目录中初始化 Git：
  
   git init && git add -A && git commit -m "initial commit"

安装完后你再手动加两个 2025 标配（1分钟搞定）：
# 进入项目
cd chiwen-web

# 强烈推荐换成 Naive UI（比 Element Plus 好看100倍）
npm install naive-ui
npm install element-plus     
npm install @element-plus/icons-vue
npm install -D less

# 图标用 lucide（超轻量好看）
npm install lucide-vue-next

# 登录流程
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


