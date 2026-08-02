# Nacos 配置快照

这里保存从本地 Nacos debug Group 提取的脱敏 YAML 模板。服务运行时仍然从 Nacos 读取配置，这些文件不会被程序自动加载。

导入方式：

1. 启动 mysql 和 nacos：docker compose up -d mysql nacos。
2. 打开 http://127.0.0.1:8848/nacos。
3. 使用对应 namespace、debug Group 和 Data ID 新建或更新配置。
4. 导入前替换 YAML 中的敏感字段占位符，尤其是数据库密码、JWT 签名、短信密钥和支付宝密钥。

## Namespace 与 Data ID

| 业务 | 环境变量 | Namespace | Data ID |
| --- | --- | --- | --- |
| 用户 | SHOP_NACOS_NAMESPACE_USERS | b5eb241f-0108-41bd-acac-ae63e1d191ae | user-srv、user-web |
| 商品 | SHOP_NACOS_NAMESPACE_GOODS | c7ad7161-42ec-4bf7-8b4f-afcd3ae58823 | goods-srv、goods-web |
| 库存 | SHOP_NACOS_NAMESPACE_INVENTORY | 801785a0-5efb-496a-9f6c-c1d7778f9642 | inventory-srv |
| 用户操作 | SHOP_NACOS_NAMESPACE_USEROP | 0d25e9b8-5908-43fe-81d1-2dff16b1ca8f | userop-srv、userop-web |
| 订单 | SHOP_NACOS_NAMESPACE_ORDER | 29e20acc-22d4-4c96-b552-f8751d32e82b | order-srv、order-web |

Group 规则：SHOP_DEBUG=true 使用 debug，false 使用 pro。

## 文件

- debug/users/user-srv.yaml
- debug/users/user-web.yaml
- debug/goods/goods-srv.yaml
- debug/goods/goods-web.yaml
- debug/inventory/inventory-srv.yaml
- debug/userop/userop-srv.yaml
- debug/userop/userop-web.yaml
- debug/order/order-srv.yaml
- debug/order/order-web.yaml

## 安全说明

文件中的 <SHOP_*> 只是占位符，不是可直接用于生产的凭据。不要把真实 Nacos 导出内容、私钥、数据库密码或 JWT 签名提交到 Git；生产环境应使用密钥管理系统或 Nacos 的受控权限管理。
