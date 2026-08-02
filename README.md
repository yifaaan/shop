# Shop

Shop 是一个面向生鲜电商场景的微服务项目，后端使用 Go，前端使用 Vue 3，服务配置由 Nacos 管理，服务发现使用 Consul，HTTP 流量通过 Traefik Catalog 自动路由。

## 项目结构

~~~text
shop/
├─ frontend/online-store/    Vue 3 + Vite 前台商城
├─ services/
│  ├─ user_srv/              用户 gRPC 服务
│  ├─ user_web/              用户 HTTP 服务
│  ├─ goods_srv/             商品 gRPC 服务
│  ├─ goods_web/             商品 HTTP 服务
│  ├─ inventory_srv/         库存 gRPC 服务
│  ├─ order_srv/             订单 gRPC 服务
│  ├─ order_web/             订单、购物车、支付 HTTP 服务
│  ├─ userop_srv/            收藏、地址、留言 gRPC 服务
│  └─ userop_web/            收藏、地址、留言 HTTP 服务
├─ pkg/                      公共库、protobuf、Nacos 配置加载
├─ docker/                   Traefik、RocketMQ、Elasticsearch 配置
├─ sql/                      MySQL 初始化脚本
├─ config/nacos/             脱敏后的 Nacos 配置快照
├─ docker-compose.yml        本地基础设施编排
└─ go.work                   Go workspace
~~~

## 架构

~~~text
Browser
   │
   ├─ Vite :4173 ──► Traefik :18000
   │                       │
   │                       └─ Consul Catalog ──► *_web ──► *_srv
   │
   ├─ Nacos :8848          配置中心
   ├─ Consul :18500        服务注册与发现
   ├─ MySQL :3306          业务数据库
   ├─ Redis :16379         缓存与验证码
   ├─ RocketMQ :18081      Proxy gRPC
   ├─ Elasticsearch :9200  商品搜索
   └─ Jaeger :16686        链路追踪
~~~

## 环境要求

- Docker Desktop
- Go 1.26.5 或兼容版本
- Node.js 18+ 和 npm
- PowerShell 7+（Windows 开发环境推荐）
- 只有修改 .proto 文件时才需要 protoc、protoc-gen-go 和 protoc-gen-go-grpc

## 快速启动

### 1. 准备环境变量

所有后端服务都应该从仓库根目录启动，因为服务会从当前目录加载 .env.local 和 .env。

~~~powershell
Set-Location D:\Code\shop

if (-not (Test-Path .env.local)) {
  Copy-Item .env.example .env.local
}
~~~

.env.local 已被 Git 忽略。请在其中填写本机覆盖项和敏感信息，不要提交真实密码、JWT 签名或第三方支付密钥。

### 2. 启动基础设施

~~~powershell
docker compose up -d
docker compose ps
~~~

首次启动会初始化 MySQL、Nacos、Redis、RocketMQ、Consul、Traefik、Elasticsearch 和 Jaeger。若已有数据卷，初始化 SQL 不会再次执行。

### 3. 检查 Nacos 配置

打开 http://127.0.0.1:8848/nacos。当前本地命名空间和 Data ID 的对应关系见 config/nacos/README.md。仓库中的 YAML 是脱敏快照，导入 Nacos 前必须替换占位符。

服务默认读取 debug Group；当 SHOP_DEBUG=false 时读取 pro Group。

### 4. 启动后端

每个命令在一个独立的 PowerShell 窗口中执行，并保持窗口运行：

~~~powershell
Set-Location D:\Code\shop

go run ./services/user_srv
go run ./services/goods_srv
go run ./services/inventory_srv
go run ./services/order_srv
go run ./services/userop_srv

go run ./services/user_web
go run ./services/goods_web
go run ./services/order_web
go run ./services/userop_web
~~~

推荐先启动所有 *_srv，确认它们已注册到 Consul，再启动 *_web。服务注册成功后，Traefik 会根据 Consul Catalog 标签动态发现 HTTP 服务。

### 5. 启动前端

~~~powershell
Set-Location D:\Code\shop\frontend\online-store
npm ci

# 默认值已经是 http://127.0.0.1:18000
$env:VITE_API_GATEWAY = "http://127.0.0.1:18000"
npm run dev -- --host 127.0.0.1
~~~

浏览器访问：http://127.0.0.1:4173/home

## 端口

### 网关和基础设施

| 组件 | 地址 | 用途 |
| --- | --- | --- |
| Vite | 127.0.0.1:4173 | 前端开发服务器 |
| Traefik | 127.0.0.1:18000 | Shop API 网关 |
| Traefik Dashboard | 127.0.0.1:18001 | 本地路由查看 |
| Nacos | 127.0.0.1:8848 | HTTP API 和控制台 |
| Nacos gRPC | 127.0.0.1:9848 | Nacos 2.x 客户端连接 |
| Consul | 127.0.0.1:18500 | HTTP API 和 Web UI |
| MySQL | 127.0.0.1:3306 | 业务数据库 |
| Redis | 127.0.0.1:16379 | 宿主机映射端口 |
| RocketMQ NameServer | 127.0.0.1:9876 | NameServer |
| RocketMQ Proxy | 127.0.0.1:18081 | Go v5 客户端 gRPC 入口 |
| Elasticsearch | 127.0.0.1:9200 | 商品搜索 |
| Kibana | 127.0.0.1:5601 | ES 管理页面 |
| Jaeger | http://127.0.0.1:16686 | Trace 查询 |

### HTTP API 前缀

| 服务 | API 前缀 | 说明 |
| --- | --- | --- |
| user_web | /u/v1 | 登录、注册、验证码、用户 |
| goods_web | /g/v1 | 商品、分类、品牌、轮播图 |
| order_web | /o/v1 | 购物车、订单、支付回调 |
| userop_web | /uo/v1 | 收藏、收货地址、留言 |

前端只需要访问 Traefik 地址，例如 http://127.0.0.1:18000/g/v1/goods。开发服务器会把 /u、/g、/o、/uo 代理到同一个网关。

## 配置约定

- 后端通过 Nacos 加载 YAML，Data ID 与服务名一致，例如 goods-srv、goods-web。
- SHOP_DEBUG=true 选择 Nacos 的 debug Group，false 选择 pro Group。
- SHOP_NACOS_HOST、SHOP_NACOS_PORT、SHOP_NACOS_USERNAME、SHOP_NACOS_PASSWORD 控制 Nacos 客户端连接。
- SHOP_NACOS_NAMESPACE_USERS、SHOP_NACOS_NAMESPACE_GOODS、SHOP_NACOS_NAMESPACE_INVENTORY、SHOP_NACOS_NAMESPACE_USEROP、SHOP_NACOS_NAMESPACE_ORDER 选择各业务命名空间。
- consul.host 和 consul.port 是宿主机上的 Consul API 地址，当前本地通常为 127.0.0.1:18500。
- consul.address 是注册给 Docker 内 Traefik 访问的宿主机地址，当前 Windows Docker 环境使用 host.docker.internal。
- SHOP_PORT 可覆盖服务监听端口。SHOP_REDIS_PORT 在宿主机开发环境通常应为 16379，容器内部端口仍为 6379。

## 测试与构建

根目录是 Go workspace，不是单独的 Go module，因此不要在根目录直接执行 go test ./...。请在各 module 目录中执行：

~~~powershell
Set-Location D:\Code\shop\pkg
go test ./...

Set-Location D:\Code\shop\services\userop_srv
go test ./...

Set-Location D:\Code\shop\services\goods_web
go test ./...

Set-Location D:\Code\shop\services\order_web
go test ./...
~~~

前端检查：

~~~powershell
Set-Location D:\Code\shop\frontend\online-store
npm run typecheck
npm run build
~~~

修改 protobuf 后生成代码：

~~~powershell
Set-Location D:\Code\shop
.\scripts\gen_proto.ps1
~~~

## 常见问题

### Nacos 配置读取失败

确认服务从 D:\Code\shop 启动，确认 Nacos 的 HTTP 端口是 8848，Nacos 2.x 客户端还需要可访问 9848。然后检查 .env.local 中的 namespace、用户名和密码。

### Traefik 找不到后端

确认 Web 服务已注册到 Consul，并且服务配置中的 consul.address 为 host.docker.internal。Traefik 容器内访问宿主机服务不能使用宿主机进程的 127.0.0.1。

### 端口冲突

当 SHOP_DEBUG=true 时，服务会使用 Nacos 中的固定端口；修改对应 Data ID 的 port 或通过 SHOP_PORT 临时覆盖。修改后重启服务，使 Consul 中的旧实例注销。

### 清理本地依赖

~~~powershell
docker compose down
~~~

docker compose down -v 会删除 MySQL、Redis、Consul、Elasticsearch 等数据卷，只应在明确需要重置本地数据时使用。
