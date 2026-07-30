--
-- Current Database: `shop_inventory_srv`
--
-- 库说明：inventory 服务专用库，与 shop_goods_srv 物理隔离。
-- 库名取自 Nacos 中 inventory-srv 配置的 mysql.dbname（约定与 shop_goods_srv 对齐）。
-- 如实际库名不同，请替换下面所有 `shop_inventory_srv` 与脚本末尾跨库种子里的引用。
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `shop_inventory_srv` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci */;

USE `shop_inventory_srv`;

--
-- Table structure for table `inventory`
--
-- 与 services/inventory_srv/inventory.go 中的 Inventory 模型一致；
-- inventory_srv 启动时也会 AutoMigrate 自动创建本表，此处给出等价的建表语句便于离线初始化/文档化。
-- 字段对应：
--   BaseModel.ID        -> id          (int, 主键, 自增)
--   BaseModel.CreatedAt -> add_time    (datetime(3))
--   BaseModel.UpdatedAt -> update_time (datetime(3))
--   BaseModel.DeletedAt -> deleted_at  (datetime(3), 软删索引)
--   Inventory.GoodsID   -> goods_id    (int, 非空, 唯一索引)
--   Inventory.Stocks    -> stocks      (int, 非空, 默认 0)
--   Inventory.Version   -> version     (int, 非空, 默认 0; 分布式乐观锁)
--

DROP TABLE IF EXISTS `inventory`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inventory` (
  `id` int NOT NULL AUTO_INCREMENT,
  `add_time` datetime(3) DEFAULT NULL,
  `update_time` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `goods_id` int NOT NULL,
  `stocks` int NOT NULL DEFAULT '0',
  `version` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq_inventory_goods_id` (`goods_id`) USING BTREE,
  KEY `idx_inventory_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for order-level inventory deductions.
-- status: 1 = deducted, 2 = returned.
--

DROP TABLE IF EXISTS `stock_sell_detail_item`;
DROP TABLE IF EXISTS `stock_sell_detail`;
CREATE TABLE `stock_sell_detail` (
  `id` int NOT NULL AUTO_INCREMENT,
  `add_time` datetime(3) DEFAULT NULL,
  `update_time` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `order_sn` varchar(40) NOT NULL,
  `status` tinyint NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq_stock_sell_detail_order_sn` (`order_sn`) USING BTREE,
  KEY `idx_stock_sell_detail_deleted_at` (`deleted_at`),
  CONSTRAINT `chk_stock_sell_detail_status` CHECK (`status` IN (1, 2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;

CREATE TABLE `stock_sell_detail_item` (
  `id` int NOT NULL AUTO_INCREMENT,
  `add_time` datetime(3) DEFAULT NULL,
  `update_time` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `stock_sell_detail_id` int NOT NULL,
  `goods_id` int NOT NULL,
  `num` int NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq_stock_sell_detail_goods` (`stock_sell_detail_id`, `goods_id`) USING BTREE,
  KEY `idx_stock_sell_detail_item_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;

--
-- Dumping data for table `inventory`
--
-- 根据 shop_goods_srv.goods 批量生成库存：
--   * 每个未软删的商品 (deleted_at IS NULL) 对应一条库存记录；
--   * stocks 取 0~2000 的随机值（FLOOR(RAND()*2001)）；
--   * version 初始化为 0；
--   * 用 INSERT IGNORE 保证可重复执行：已存在的 goods_id 不会被覆盖，仅补齐缺失的商品。
--
-- 前置条件：
--   1. shop_goods_srv 库已存在且 goods 表已有数据（通常先启动 goods_srv 导入 01-goods-srv.sql）；
--   2. 执行账号需同时具备 shop_goods_srv 的 SELECT 权限与 shop_inventory_srv 的 INSERT 权限
--      （docker-compose 中 shop_user 默认仅授权单库，跨库种子请用 root 或赋权后的账号执行）。
--

LOCK TABLES `inventory` WRITE;
/*!40000 ALTER TABLE `inventory` DISABLE KEYS */;
INSERT IGNORE INTO `inventory` (`goods_id`, `stocks`, `version`, `add_time`, `update_time`)
SELECT g.`id`, FLOOR(RAND() * 2001), 0, NOW(3), NOW(3)
FROM `shop_goods_srv`.`goods` g
WHERE g.`deleted_at` IS NULL;
/*!40000 ALTER TABLE `inventory` ENABLE KEYS */;
UNLOCK TABLES;

-- 预期：执行后 inventory 记录数 == shop_goods_srv.goods 中未软删的商品数。
-- 可用以下语句核对：
--   SELECT COUNT(*) AS goods_cnt  FROM `shop_goods_srv`.`goods`    WHERE deleted_at IS NULL;
--   SELECT COUNT(*) AS inv_cnt    FROM `shop_inventory_srv`.`inventory`;
