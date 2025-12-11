客户端启动 首先检查 用户 是否存在RSA 公私钥对
如果没有生成
    nonce随机数（随机64字节）
    当前时间戳（秒级） 
    uuid 
    生成 RSA 公私钥对
    计算签名 signature 
使用客户端的私钥对这个数据包进行签名
将原始数据包 + 签名一起发送给服务端
{
    "nonce": "随机64字节Base64编码",
    "timestamp": 1700000000,
    "uuid": "客户端唯一标识",
    "public_key": "客户端RSA公钥PEM格式",
    "signature": "签名字符串Base64编码"
    “hotname”：“xxx”
}

服务端：
校验 timestamp（±120 秒）
校验 signature（RSA-Verify）  
查询 client_nonce 是否重复   
将注册申请记录写入 assets_apply  

管理员在后台进行审批 通过后 写入正式资产表 assets   
这个地方应该有个前端 然后管理员点击按钮 就通过了审批 后面再做这个前端  
通过审批时
    获取 临时注册表 中的 uuid client_public_key 注册时间 主机名
    由server端生成的 
        状态 online 
        labels 是空
        updated 心跳时间 应该怎么写？或者说是写什么数据？逻辑是如何的？
        数据软删除 deleted 是空
写入后 正式资产表 
服务端 生成 agent_secret_key 
会用客户端的公钥（client_public_key）进行加密，然后发送给客户端

客户端
Client 使用 client_private_key 解密 agent_secret_key
并 保存 agent_secret_key  然后agent_secret_key 存储在 ~/.ssh 目录下
Client 开始心跳 （心跳是客户端定期向服务端发送状态报告）
{
  "id": "xxxx",
  "timestamp": 1700000000,
  "metrics": {
	动态信息
    	"cpu_usage": 45.2,    // cpu 使用率
    	"memory_usage": 68.5, // 内存使用率
    	"disk_usage": 32.1,   // 磁盘使用率 如果是多块磁盘 如何处理？
		网络网卡接口速度： 接口速度（Mbps）	
	静态信息
		主机名
		操作系统信息
			系统类型： 是ubuntu 还是centos +版本
			内核+版本
			系统启动时间：
		操作系统硬件信息
			cpu 个数
			cpu 频率
			磁盘 个数 
			磁盘 类型
			磁盘 大小
			内存 大小
			交换分区 大小
		网络信息
			网卡名称：eth0 
			内网地址ipv4 
			公网地址ipv4

		
  },
  "signature": "使用agent_secret_key签名的Base64"
}

每次 client 客户端 启动 后 静态信息（只在第一次或变化时发送）
动态信息（每次心跳都发送）
心跳间隔30秒 

服务端收到心跳：
校验 timestamp（±120s）
校验签名（使用 agent_secret_key）
更新动态数据 
如果静态数据有变化，覆盖
更新心跳时间 assets.updated_at 
心跳正常 → 返回简单 OK




心跳异常处理机制 没有心跳超过 1 分钟 设为 offline（updated_at > 3 分钟 → offline） 这个步骤应该放哪里？









问题：
go run cmd/main.go 
{"level":"INFO","time":"2025-12-06T14:00:22.393Z","caller":"app/client.go:54","msg":"Logger initialized"}
{"level":"INFO","time":"2025-12-06T14:00:22.393Z","caller":"handler/register_handler.go:42","msg":"client id ready","id":"758722b9-84a3-4de8-ab35-4b32f8e32ac3"}
{"level":"INFO","time":"2025-12-06T14:00:22.393Z","caller":"handler/register_handler.go:49","msg":"rsa keys ready","pub_len":"451"}
{"level":"ERROR","time":"2025-12-06T14:00:22.396Z","caller":"app/client.go:61","msg":"register failed","error":"send register error: server returned 400: {\"error\":\"insert apply failed: Error 1062 (23000): Duplicate entry '758722b9-84a3-4de8-ab35-4b32f8e32ac3' for key 'agent_register_apply.PRIMARY'\"}"}
Error: send register error: server returned 400: {"error":"insert apply failed: Error 1062 (23000): Duplicate entry '758722b9-84a3-4de8-ab35-4b32f8e32ac3' for key 'agent_register_apply.PRIMARY'"}










按照流程-------------如下
客户端启动 首先检查 用户 是否存在RSA 公私钥对
如果没有生成
    nonce随机数（随机64字节）
    当前时间戳（秒级） 
    uuid 
    生成 RSA 公私钥对
    计算签名 signature 
使用客户端的私钥对这个数据包进行签名
将原始数据包 + 签名一起发送给服务端
{
    "nonce": "随机64字节Base64编码",
    "timestamp": 1700000000,
    "uuid": "客户端唯一标识",
    "public_key": "客户端RSA公钥PEM格式",
    "signature": "签名字符串Base64编码"
    “hotname”：“xxx”
}

服务端：
校验 timestamp（±120 秒）
校验 signature（RSA-Verify）  
查询 client_nonce 是否重复   
将注册申请记录写入 assets_apply  

管理员在后台进行审批 通过后 写入正式资产表 assets   
这个地方应该有个前端 然后管理员点击按钮 就通过了审批 后面再做这个前端  
通过审批时
    获取 临时注册表 中的 uuid client_public_key 注册时间 主机名
    由server端生成的 
        状态 online 
        labels 是空
        updated 心跳时间 应该怎么写？或者说是写什么数据？逻辑是如何的？
        数据软删除 deleted 是空
写入后 正式资产表 
服务端 生成 agent_secret_key 
会用客户端的公钥（client_public_key）进行加密，然后发送给客户端


客户端
Client 使用 client_private_key 解密 agent_secret_key
并 保存 agent_secret_key  然后agent_secret_key 存储在 ~/.ssh 目录下
Client 开始心跳 （心跳是客户端定期向服务端发送状态报告）
{
  "id": "xxxx",
  "timestamp": 1700000000,
  "metrics": {
	动态信息
    	"cpu_usage": 45.2,    // cpu 使用率
    	"memory_usage": 68.5, // 内存使用率
    	"disk_usage": 32.1,   // 磁盘使用率 如果是多块磁盘 如何处理？
		网络网卡接口速度： 接口速度（Mbps）	
	静态信息
		主机名
		操作系统信息
			系统类型： 是ubuntu 还是centos +版本
			内核+版本
			系统启动时间：
		操作系统硬件信息
			cpu 个数
			cpu 频率
			磁盘 个数 
			磁盘 类型
			磁盘 大小
			内存 大小
			交换分区 大小
		网络信息
			网卡名称：eth0 
			内网地址ipv4 
			公网地址ipv4

		
  },
  "signature": "使用agent_secret_key签名的Base64"
}

每次 client 客户端 启动 后 静态信息（只在第一次或变化时发送）
动态信息（每次心跳都发送）
心跳间隔30秒 

服务端收到心跳：
校验 timestamp（±120s）
校验签名（使用 agent_secret_key）
更新动态数据 
如果静态数据有变化，覆盖
更新心跳时间 assets.updated_at 
心跳正常 → 返回简单 OK
