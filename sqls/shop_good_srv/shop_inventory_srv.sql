/*
 Navicat Premium Dump SQL

 Source Server         : shop_user
 Source Server Type    : MySQL
 Source Server Version : 80044 (8.0.44)
 Source Host           : localhost:3306
 Source Schema         : shop_inventory_srv

 Target Server Type    : MySQL
 Target Server Version : 80044 (8.0.44)
 File Encoding         : 65001

 Date: 02/02/2026 20:12:58
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for inventory
-- ----------------------------
DROP TABLE IF EXISTS `inventory`;
CREATE TABLE `inventory`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `add_time` datetime(3) NULL DEFAULT NULL,
  `is_deleted` int NULL DEFAULT 0 COMMENT '\'是否删除\'',
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `good` int NOT NULL COMMENT '\'商品ID\'',
  `stock` int NOT NULL COMMENT '\'库存数量\'',
  `version` int NOT NULL COMMENT '\'版本号\'',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_inventory_deleted_at`(`deleted_at` ASC) USING BTREE,
  UNIQUE INDEX `idx_inventory_good`(`good` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of inventory
-- ----------------------------

SET FOREIGN_KEY_CHECKS = 1;
