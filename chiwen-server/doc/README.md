开发流程
1 首先看业务 想实现的功能
2 梳理业务流程图
3 抽象功能完成 表设计 
4 搭建脚手架
5 routes --- handler（Controller）--- service（logic）--- data（dao）
			请求结构体									 持久结构体


创建数据库（版本：8.0.44）
docker run -d --name mysql8 --hostname mysql-server --restart=unless-stopped \
-e MYSQL_ROOT_PASSWORD=R00tP@ssw0rd! \
-e MYSQL_DATABASE=myapp \
-e MYSQL_USER=myuser \
-e MYSQL_PASSWORD=MyUserP@ss123 \
-e TZ=Asia/Shanghai \
-p 3306:3306 \
-v /data/mysql/server_data:/var/lib/mysql \
-v /data/mysql/conf:/etc/mysql/conf.d \
-v /data/mysql/logs:/var/log/mysql \
mysql:8.0 \
--character-set-server=utf8mb4 \
--collation-server=utf8mb4_unicode_ci \
--default-authentication-plugin=mysql_native_password \
--max_connections=500 \
--innodb_buffer_pool_size=512M

创建数据表
CREATE TABLE `assets` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机器唯一ID',
  `client_public_key` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '客户端RSA公钥',
  `hostname` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机器主机名',
  `labels` json DEFAULT NULL COMMENT '机器标签，JSON格式存储',
  `allowed_users` json DEFAULT NULL COMMENT '允许直接连接此机器的用户ID列表，示例：["1","3","7"]',
  `status` enum('online','offline','maintenance') COLLATE utf8mb4_unicode_ci DEFAULT 'offline' COMMENT '机器状态',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '机器注册/发现时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '心跳时间/更新时间',
  `is_deleted` tinyint(1) DEFAULT '0' COMMENT '软删除标记',
  `static_info` json DEFAULT NULL COMMENT '静态信息（CPU/OS/磁盘/网卡等）',
  `dynamic_info` json DEFAULT NULL COMMENT '动态信息（CPU使用率/内存/磁盘使用率等）',
  `agent_secret_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用于心跳HMAC的secret',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_hostname` (`hostname`),
  KEY `idx_updated_at` (`updated_at`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_is_deleted` (`is_deleted`),
  KEY `idx_allowed_users` ((cast(`allowed_users` as unsigned array)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='正式机器资产表'

CREATE TABLE `agent_register_apply` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nonce` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '客户端随机数（防重放）',
  `hostname` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主机名',
  `apply_status` enum('pending','approved','rejected') COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '申请状态',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请创建时间',
  `client_public_key` text COLLATE utf8mb4_unicode_ci COMMENT '客户端 RSA 公钥',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_id` (`id`),
  KEY `idx_status` (`apply_status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent 注册申请表'

CREATE TABLE `tty_sessions` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话ID',
  `asset_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机器ID',
  `user_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户ID',
  `token` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '一次性Token',
  `status` enum('pending','connected','closed','error') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '会话状态',
  `command` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '/bin/bash' COMMENT '执行的命令',
  `terminal_cols` int DEFAULT '80' COMMENT '终端列数',
  `terminal_rows` int DEFAULT '24' COMMENT '终端行数',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `connected_at` timestamp NULL DEFAULT NULL COMMENT '连接时间',
  `closed_at` timestamp NULL DEFAULT NULL COMMENT '关闭时间',
  `browser_ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '浏览器IP',
  `agent_ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Agent IP',
  `record_file` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '录像文件路径',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_asset_id` (`asset_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_token_expire` (`token`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='TTY会话表'


脚手架
mkdir project
cd project
mkdir server client
cd server && go mod init github.com/chiwen/server

该项目布局
server
├── cmd
│   ├── app
│   └── main.go
├── configs
│   └── config.yaml
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   ├── handler			# （业务处理）
│   │   └── routes			# （路由层）
│   │       └── routes.go
│   ├── data
│   │   ├── model
│   │   └── mysql
│   │       └── mysql.go
│   └── service
├── pkg
│   ├── config
│   │   └── config.go
│   └── logger
│       └── logger.go
├── README.md

标准 Go 项目布局
your_project/
├── cmd                     # main文件统一放在/cmd目录下
internal/             		# 私有应用和库代码
├── api/              		# HTTP接口层（相当于Controller）服务的入口 负责处理路由 参数校验 请求转发
│   ├── handler/      		# 请求处理器
│   ├── middleware/   		# 中间件
│   └── router.go     		# 路由定义
├── service/          		# 业务逻辑层（相当于logic）负责处理业务逻辑
│   ├── user/
│   └── product/
├── data/             		# 数据访问层（相当于DAO）负责数据与存储相关功能
│   ├── dao/          		# 数据访问对象
│   └── model/        		# 数据模型
└── pkg/              		# 内部工具包（仅本项目使用）
    ├── util/
    └── constant/            
├── pkg                     # 存放可以被外部应用使用的代码库
├── api                     # OpenAPI/Swagger
├── configs                 # 配置文件模板或默认配置
├── README.md               # 项目的介绍
└── tools                   # 这个项目的支持工具
├── go.mod
└── go.sum

├── web/
├── docs                    # 设计和用户文档
├── githooks                # Git钩子
├── third_party             # 外部帮助工具分支代码
├── examples                # 存放应用程序或者公共包的示例代码
├── scripts                 # 各类脚本文件
├── init                    # 存放始化系统和进程管理配置文件
├── deployments             # 容器编排部署配置和模板
├── build                   # 包和持续集成相关文件
├── test                    # 其他外部测试应用和测试数据
├── LICENSE                 # 版权文件
├── CONTRIBUTING.md         # 贡献指南文件
├── CHANGELOG               # 版本变更历史
├── Makefile                # 项目管理工具

业务流程

Agent 注册到上线全流程
    客户端启动
        1 检查 ~/.ssh 目录是否存在，不存在则创建
        2 检查 client_id 文件 → 无则生成 UUID 并写入
        3 检查 RSA 密钥对 → 无则生成 2048 位密钥对
        4 检查 agent_secret_key 文件 → 存在则直接跳到第12步开始心跳
        5 构造注册请求生成随机 nonce + timestamp + hostname + 读取公钥
          POST /api/v1/register
    服务端
        6 验 timestamp (±120s)
          验 RSA 签名
          验 nonce 不重复
          若已有 pending 申请则直接返回 pending
          否则插入agent_register_apply
          (数据库插入 agent_register_apply
            id = uuid
            apply_status = "pending"
          返回 {"status":"pending","apply_id":"uuid"}
    客户端
        7 收到 pending → 每 10s 轮询一次 /api/v1/register/status?apply_id=uuid
          GET /register/status
    管理员
        8 用 Insomnia/Postman 调用审批接口
          POST /api/v1/approve {"id":"uuid"}
    服务端
        9 查询 agent_register_apply（任意状态）
          生成 32 字节随机 secret → base64 → secretStr
          用客户端公钥加密 secretStr → encryptedSecret
          REPLACE INTO assets（防止冲突）
          UPDATE agent_register_apply.apply_status = "approved" 不再删除申请记录
          返回 {"status":"approved","encrypted_secret":"..."}
          (数据库 assets 表插入/替换一行
          (agent_register_apply.apply_status 改为 approved
        10 客户端继续轮询时：
           先查 agent_register_apply，发现 apply_status="approved"
           再从 assets 表取明文 secret，用客户端公钥重新加密返回
           返回 {"status":"approved","encrypted_secret":"xxxx"}
    客户端
        11 收到 encrypted_secret → 用本地私钥解密 → 得到明文 agent_secret_key → 写入文件
           注册成功！
        12 启动心跳循环（每 30s 一次）构造 payload = `id
           关键请求/响应：timestamp
           (metricsJSON 用 agent_secret_key 做 HMAC-SHA256 → signature
    服务端
        13 心跳处理：
           验 timestamp (±120s)
           从 assets 表取agent_secret_key
           验 HMAC 签名
           更新资产信息和心跳时间
           返回 {"status":"ok"}
           (assets 表更新


tty 逻辑
Web终端代理架构，允许用户通过浏览器访问服务器上的终端。架构分为三部分：
  浏览器端：用户通过WebSocket连接到服务器
  服务端：作为中继和权限控制中心
  Agent端：部署在目标机器上，负责建立实际的终端连接

1 浏览器发起WebSocket连接
  用户点击Web界面 → 浏览器建立WebSocket连接
  wss://your.com/api/v1/tty/ws?id=xxxx
  携带机器ID和用户凭证

2 服务端权限验证
  通过目标机器ID查询assets表，验证用户是否有权限访问该机器
  生成一次性token（防泄露）
  返回token给客户端

3 客户端client轮询获取任务
  client 心跳上报时会轮询 /api/v1/agent/tty/sessions
  每5秒轮询一次
  查询是否有自己的TTY任务
  有任务，获取token和目标信息
  
4 连接到服务器的中继服务 （下面的这些动作client 操作的）
    建立WebSocket连接到中继
    在本地创建PTY（伪终端）
    启动Shell（或自定义命令）
    双向数据转发
      WebSocket → PTY
      PTY → WebSocket

5 中继服务连接两端（服务器）
  浏览器连接
    // 验证token
    // 等待Agent连接
    // Agent已连接，开始转发
  Agent连接
    // 通知浏览器端Agent已就绪


# WebTTY 完整流程（三方协作版：浏览器 ↔ Server ↔ Agent）

## 1. 用户在前端点击「打开终端」
   前端 → GET /api/v1/assets/{id}/tty/authorize
   后端校验：
     - 机器存在且 status='online' （是否存在且状态正常）
     - 用户在该资产的「授权用户列表」中（通过任意一种简单绑定方式）
     - 当前并发会话数未超限
     - 接口频率限制 防暴力枚举资产ID）
     Token 防重放与时效性
   成功 → 返回 JSON：
     {
       "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",   // 5分钟有效一次性token
       "ws_url": "wss://your.com/api/v1/tty/ws"
     }

已测试：
INSERT INTO assets (id, client_public_key, hostname, allowed_users, status, agent_secret_key) 
VALUES (
  'test-machine-001',
  '-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyour_test_public_key_here\n-----END PUBLIC KEY-----',
  'test-server-01',
  '["1", "2", "3"]',  -- 允许用户1,2,3访问
  'online',
  'test_agent_secret_key_123456'
);

未测试
INSERT INTO assets (id, client_public_key, hostname, allowed_users, status, agent_secret_key) 
VALUES (
  'test-machine-002',
  '-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAanother_key\n-----END PUBLIC KEY-----',
  'test-server-02',
  '["99", "100"]',  -- 只允许用户99和100访问
  'online',
  'another_agent_secret_key'
);

未测试
INSERT INTO assets (id, client_public_key, hostname, allowed_users, status, agent_secret_key) 
VALUES (
  'test-machine-offline',
  '-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoffline_key\n-----END PUBLIC KEY-----',
  'test-server-offline',
  '["1", "2", "3"]',
  'offline',  -- 离线状态
  'offline_agent_secret_key'
);

curl -v "http://192.168.19.100:8090/api/v1/assets/test-machine-001/tty/authorize?user_id=1&cols=100&rows=30"
*   Trying 192.168.19.100:8090...
* Connected to 192.168.19.100 (192.168.19.100) port 8090 (#0)
> GET /api/v1/assets/test-machine-001/tty/authorize?user_id=1&cols=100&rows=30 HTTP/1.1
> Host: 192.168.19.100:8090
> User-Agent: curl/7.81.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Mon, 08 Dec 2025 08:58:09 GMT
< Content-Length: 173
< 
* Connection #0 to host 192.168.19.100 left intact
{"expires_in":300,"session":{"cols":100,"command":"/bin/bash","rows":30},"token":"7jtkVggFblzCsyAjz0uZPg0oVXpVXhjXzqxYeQPRfbs=","ws_url":"ws://localhost:8090/api/v1/tty/ws"}

apt install node-ws
wscat -c "ws://192.168.19.100:8090/api/v1/tty/ws?token=7jtkVggFblzCsyAjz0uZPg0oVXpVXhjXzqxYeQPRfbs="
error: Unexpected server response: 404






## 2. 浏览器建立 WebSocket（用户侧）
   前端用 xterm.js + gorilla/websocket 风格：
   ws = new WebSocket(`wss://your.com/api/v1/tty/ws?token=${token}&cols=180&rows=40`);
   连接成功后，浏览器只管：
     - 键盘输入 → ws.send(data)
     - ws.onmessage → term.write(data)
     - ws.onclose → 提示会话结束

## 3. 服务端收到浏览器 WebSocket（/api/v1/tty/ws）
   Gin Handler 做的事：
     - 校验 token → 解析出 asset_id + user_id + user_name
     - 生成 session_id（uuid）
     - 在数据库 tty_sessions 表插入一条记录（status='connecting'）
     - 把这个浏览器 WebSocket 存入全局 map：
         SessionPool[session_id] = &BrowserConn{ws, cols, rows}
     - 给前端回一个 {"session_id": "..."} 表示就绪
     - 现在服务端只等 Agent 来“认领”这个会话

## 4. Agent 轮询拿到任务（关键！这一步是 Agent 主动做的）
   Agent 每 5~8 秒轮询一次（和心跳一起就行）：
   GET /api/v1/agent/tty/pending-sessions?agent_id=xxx
   服务端返回：
     []{
       "session_id": "a1b2c3...",
       "token": "xxxx",
       "cols": 180,
       "rows": 40,
       "user": "admin"
     }
   Agent 拿到后立即：
     - 校验 token 有效且属于自己的机器
     - 主动建立 WebSocket 连回服务器（这就是“反向连接”的精髓）

## 5. Agent 主动连接服务器中继（Agent → Server）
   Agent 用 websocket.Dialer 连接：
   wss://your.com/api/v1/agent/tty/relay?session_id=a1b2c3...&token=xxxx
   连接成功后，Agent 立刻在本地干三件事：
     ① 用 creack/pty.StartWithSize("/bin/bash --login", cols, rows)
        （Windows 就用 conpty）
     ② 把 pty.Master 的 Stdin/Stdout/Stderr 接管
     ③ 开始双向 Copy（经典四行代码）：
         go io.Copy(ptyMaster, agentWS)    // 浏览器输入 → shell
         go io.Copy(agentWS, ptyMaster)    // shell 输出 → 浏览器

## 6. 服务端收到 Agent 的反向 WebSocket（/api/v1/agent/tty/relay）
   Gin Handler 做的事：
     - 校验 session_id + token
     - 从 SessionPool 取出对应的浏览器 WebSocket
     - 更新数据库 tty_sessions status='active', started_at=NOW()
     - 开始纯中转（四行代码）：
         go io.Copy(browserWS, agentWS)
         go io.Copy(agentWS, browserWS)
     - 同时开一个 goroutine 做录像（可选但强烈推荐）：
         teeReader := io.TeeReader(agentWS, recordingWriter) // 录下 shell 输出

## 7. 会话结束
   任意一方关闭 WebSocket → 另一方感知到 EOF → 都关闭
   服务端清理 SessionPool + 更新数据库 ended_at + 上传录像
   前端 xterm.js 显示 “Disconnected”



后期：RBAC/标签授权系统


关键组件设计
  1. 权限验证层
  // 权限检查
    // 1. 用户是否是该机器的管理员？
    // 2. 用户是否在机器标签的授权列表中？
    // 3. 是否是特定时间段？
    // 4. 是否达到并发连接限制？
  2. 会话管理
type TTYSession struct {
    ID         string    `json:"id"`
    MachineID  string    `json:"machine_id"`
    UserID     string    `json:"user_id"`
    Token      string    `json:"token"`
    Status     string    `json:"status"` // pending, connected, closed
    CreatedAt  time.Time `json:"created_at"`
    ConnectedAt time.Time `json:"connected_at"`
    BrowserIP  string    `json:"browser_ip"`
    AgentIP    string    `json:"agent_ip"`
    
    // PTY相关
    PTY        *os.File  `json:"-"`
    Command    string    `json:"command"`  // 执行的命令
    Cols       int       `json:"cols"`     // 终端列数
    Rows       int       `json:"rows"`     // 终端行数
    
    // 审计
    RecordFile string    `json:"record_file"`  // 录像文件路径
    LogFile    string    `json:"log_file"`     // 操作日志
}

录像功能
  // 使用 asciinema 格式
  // 记录所有输入输出

 Agent心跳上报包含TTY状态
  // Agent心跳数据扩展
type HeartbeatRequest struct {
    ID        string                 `json:"id"`
    Timestamp int64                  `json:"timestamp"`
    Metrics   map[string]interface{} `json:"metrics"`
    Signature string                 `json:"signature"`
    
    // TTY相关状态
    TTYStatus struct {
        ActiveSessions int      `json:"active_sessions"`
        Sessions       []string `json:"sessions"`  // 当前活跃的session IDs
    } `json:"tty_status,omitempty"`
}


CREATE TABLE `tty_sessions` (
  `id` varchar(36) NOT NULL PRIMARY KEY,
  `asset_id` varchar(64) NOT NULL,
  `user_id` int NOT NULL,
  `username` varchar(64) NOT NULL,
  `client_ip` varchar(64),
  `cols` int DEFAULT 180,
  `rows` int DEFAULT 40,
  `status` enum('connecting','active','closed') DEFAULT 'connecting',
  `token` varchar(128) UNIQUE,           -- 一次性授权 token
  `recording_url` varchar(512),          -- 录像地址
  `started_at` timestamp NULL,
  `ended_at` timestamp NULL,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_asset_status (asset_id, status),
  INDEX idx_user (user_id)
);






表结构
```markdown
CREATE TABLE `assets` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机器唯一ID',
  `client_public_key` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '客户端RSA公钥',
  `hostname` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机器主机名',
  `labels` json DEFAULT NULL COMMENT '机器标签，JSON格式存储',
  `allowed_users` json DEFAULT NULL COMMENT '允许直接连接此机器的用户ID列表，示例：["1","3","7"]',
  `status` enum('online','offline','maintenance') COLLATE utf8mb4_unicode_ci DEFAULT 'offline' COMMENT '机器状态',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '机器注册/发现时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '心跳时间/更新时间',
  `is_deleted` tinyint(1) DEFAULT '0' COMMENT '软删除标记',
  `static_info` json DEFAULT NULL COMMENT '静态信息（CPU/OS/磁盘/网卡等）',
  `dynamic_info` json DEFAULT NULL COMMENT '动态信息（CPU使用率/内存/磁盘使用率等）',
  `agent_secret_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用于心跳HMAC的secret',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_hostname` (`hostname`),
  KEY `idx_updated_at` (`updated_at`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_is_deleted` (`is_deleted`),
  KEY `idx_allowed_users` ((cast(`allowed_users` as unsigned array)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='正式机器资产表'
```

https://cowtransfer.com/s/fc5729d5abef47
口令：0p5bso
请查看 上面的代码
先帮我实现下面 server端 逻辑的代码 一步一步教我 如何实现

```markdown
# WebTTY 完整流程（三方协作版：浏览器 ↔ Server ↔ Agent）

## 1. 用户在前端点击「打开终端」
   前端 → GET /api/v1/assets/{id}/tty/authorize
   后端校验：
     - 机器存在且 status='online' （是否存在且状态正常）
     - 用户在该资产的「授权用户列表」中（通过任意一种简单绑定方式）
     - 当前并发会话数未超限
     - 接口频率限制 防暴力枚举资产ID）
     Token 防重放与时效性
   成功 → 返回 JSON：
     {
       "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",   // 5分钟有效一次性token
       "ws_url": "wss://your.com/api/v1/tty/ws"
     }
```

