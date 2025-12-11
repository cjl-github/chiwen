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


前端 (浏览器)                    后端 (Server)                   后端 (Agent)
     |                                |                               |
     | 1. GET /authorize              |                               |
     | user_id, cols, rows ---------> |                               |
     |                                |                               |
     |                                | 2. 验证权限，生成token           |
     |                                | 写入tty_sessions               |
     |                                | (包含cols,rows)                |
     |                                |                               |
     | 3. 返回token, ws_url <-------- |                                |
     |                                |                               |
     | 4. WebSocket连接              |                                 |
     | ws://...?token=xxx ---------> |                                |
     |                                |                               |
     |                                | 5. 验证token，获取会话          |
     |                                | 查找对应Agent连接               |
     |                                |                               |
     | 6. 连接成功                  | 6. 找到Agent，配对转发             |
     | 开始渲染终端                |                                    |
     |                                | 7. 通知Agent创建PTY            |
     |                                | ----------------------------> |
     |                                |                               | 8. 创建PTY，开始Shell
     | 9. 用户输入                 |                                   |
     | ws.send(data) -------------> | 10. 转发到Agent ------------>    |
     |                                |                               | 11. 写入PTY
     |                                |                               |
     | 12. PTY输出                | 12. 转发到前端 <------------        |
     | <--------------------------- |                                 |
     | terminal.write(data)          |                                |

    ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
    │   前端Web   │─────▶│   Server    │◀────│   Agent     │
    │             │◀─────│   (后端)    │────▶│   (客户端)  │
    └─────────────┘      └─────────────┘      └─────────────┘
          │                      │                      │
          │                      │                      │
          ▼                      ▼                      ▼
    ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
    │  浏览器     │      │   数据库     │      │   Shell     │
    │  xterm.js   │      │  MySQL      │      │   PTY       │
    └─────────────┘      └─────────────┘      └─────────────┘

# WebTTY 完整流程（三方协作版：浏览器 ↔ Server ↔ Agent）
   前端用 xterm.js + gorilla/websocket 风格：
   ws = new WebSocket(`wss://your.com/api/v1/tty/ws?token=${token}&cols=180&rows=40`);
   连接成功后，浏览器只管：
     - 键盘输入 → ws.send(data)
     - ws.onmessage → term.write(data)
     - ws.onclose → 提示会话结束

## 1. 用户在前端点击「打开终端」
   前端 → GET /api/v1/assets/{id}/tty/authorize
   后端校验逻辑
     1 校验机器是否存在 查询 assets 表中 id = asset_id，如果不存在 → 返回 404
	 2 校验机器状态 status 必须是 online，若是 offline or maintenance → 返回 400 + message
     3 校验用户权限
	 用户在该资产的「授权用户列表」中（通过任意一种简单绑定方式）
     4 并发会话数限制（可选）
	 查询 tty_sessions 表中：
		asset_id = asset_id
		status in ('pending','connected') 
	 超过限制 → 429 Too Many Requests
     5 接口频率限制（可选）
		根据 user_id + asset_id 做简单的 Rate Limit
		防止暴力枚举 asset_id
     6 生成一次性 Token（用于 WebSocket 连接）
	   后端生成 token 时写入 MySQL
	   WebSocket 连接时验证 token
	   Token 必须 一次性使用（首次被 WS 消费后立即失效）
	   返回 JSON 给前端（用于建立 WebSocket 连接）
	    成功 → 返回 JSON：
       {
       "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",   // 5分钟有效一次性token
       "ws_url": "wss://your.com/api/v1/tty/ws"
       }
	   Token 规则：
		 长度至少 32 字节随机串
		 Base64 编码
	     设置 5 分钟有效期
	 7 清理过期 token（建议定时任务）

建一张独立的 token 表
CREATE TABLE tty_tokens (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    token VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    asset_id BIGINT NOT NULL,
    cols INT DEFAULT 120,
    rows INT DEFAULT 30,
    status ENUM('pending', 'used', 'expired') DEFAULT 'pending',
    expire_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME NULL,
    INDEX idx_expire_at (expire_at),
    INDEX idx_user_asset (user_id, asset_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


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

wscat -c "ws://127.0.0.1:8090/api/v1/tty/ws?token=YOUR_TOKEN"


Agent 长连接机制
  1. Agent 启动时主动连接 Server
    Agent 在启动后执行：WebSocket →  wss://server/api/v1/tty/agent/ws
    这是一个长期保持的双向 WS 连接。
  2. Agent 发起认证（asset_id + timestamp + HMAC 签名）
    Agent 连接时必须携带：
      asset_id
      timestamp（当前时间戳）
      signature = HMAC-SHA256( asset_id + timestamp , agent_secret_key )
      例子：wss://server/api/v1/tty/agent/ws?asset_id=test-machine-001&ts=1733639373&sig=abcdef123456...
  3. Server 校验步骤：
    asset_id 是否存在（查 assets）
    是否 online
    找到 asset.agent_secret_key
    验证签名：expected = HMAC(asset_id+ts, agent_secret_key)
    如果 signature != expected → 拒绝连接
    通过认证后：将 Agent WS 连接状态记录到 MySQL
  4. Agent 保持心跳
    为了保持长连接，Agent 每隔 10–30 秒发送心跳包
    Server 能识别 Agent 是否掉线
    如果 Server 超过 60 秒未收到 ping
    标记 Asset 为 offline（更新 assets.status）
    断开 WebSocket

## 2. 建立 WebSocket
    用户从 /tty/authorize 获取到一次性 token 后，前端开始建立 WebSocket：
     ws = new WebSocket(
     wss://your.com/api/v1/tty/ws?token=${token}&cols=180&rows=40
    )
    连接成功后，前端只负责：
        - 用户输入 → ws.send(data)
        - ws.onmessage → term.write(data)
        - ws.onclose → 提示用户会话结束


    GET  /api/v1/assets/{id}/tty/authorize  # 获取token
    WS   /api/v1/tty/ws                     # 用户WebSocket
    WS   /api/v1/tty/agent/ws               # Agent WebSocket


    实现后端的WebSocket服务端逻辑
		1. 升级 WebSocket
        WebSocket升级是HTTP协议到WebSocket协议的转换过程
        后端处理路径：GET /api/v1/tty/ws?token=xxxx&cols=xxx&rows=xxx
            升级为 WebSocket
            提取 token/cols/rows
        2. 校验 token
        验证内容：
            token 是否存在
            status = 'pending'
            expire_at > NOW()
        如果失败 → 立即关闭 WebSocket
        验证成功：
            更新 token → status = 'used'（防重放）
            创建/更新 tty_sessions 会话记录为 "connected"
        3. 查找 Asset 对应的 Agent WebSocket 是否在线
        后端检查：
            Agent 是否已通过 /tty/agent/ws 在线注册
            asset_id 必须匹配
            如果在线 → 建立用户WS ↔ AgentWS 双向转发
            如果不在线 → 返回错误并关闭用户 WS

        4. 用户 WS 与 Agent WS 配对（配对转发）
            用户连接 ↔ Server ↔ Agent连接
            双向转发数据
            处理resize、心跳等控制消息
        
        附加逻辑异
            长时间无输入自动关闭
            
        错误处理
            浏览器侧：ws.onclose → 显示提示即可（这个等做前端web项目等时候在做）
            后端侧：
                用户断开 → 会话关闭 → 通知 Agent 停止 pty
                Agent 断开 → 更新会话状态 + 关闭用户 WS 
                更新会话状态为 closed
                token 过期/重复使用 → 拒绝连接
        实现前端WebSocket连接和终端渲染





