# chiwen-web

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Recommended Browser Setup

- Chromium-based browsers (Chrome, Edge, Brave, etc.):
  - [Vue.js devtools](https://chromewebstore.google.com/detail/vuejs-devtools/nhdogjmejiglipccpnnnanhbledajbpd) 
  - [Turn on Custom Object Formatter in Chrome DevTools](http://bit.ly/object-formatters)
- Firefox:
  - [Vue.js devtools](https://addons.mozilla.org/en-US/firefox/addon/vue-js-devtools/)
  - [Turn on Custom Object Formatter in Firefox DevTools](https://fxdx.dev/firefox-devtools-custom-object-formatters/)

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```




```
安装 nvm
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

# 图标用 lucide（超轻量好看）
npm install lucide-vue-next
```



npm install element-plus     
npm install @element-plus/icons-vue