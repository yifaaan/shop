/*
 Navicat Premium Dump SQL

 Source Server         : shop_user
 Source Server Type    : MySQL
 Source Server Version : 80044 (8.0.44)
 Source Host           : localhost:3306
 Source Schema         : shop_user_srv

 Target Server Type    : MySQL
 Target Server Version : 80044 (8.0.44)
 File Encoding         : 65001

 Date: 02/02/2026 20:13:10
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `add_time` datetime(3) NULL DEFAULT NULL,
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `mobile` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `password` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `nick_name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `birthday` datetime NULL DEFAULT NULL,
  `gender` varchar(6) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'male' COMMENT 'female 女, male男',
  `role` int NULL DEFAULT 1 COMMENT '1表示用户, 2表示管理员',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_mobile`(`mobile` ASC) USING BTREE,
  INDEX `idx_user_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 12 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES (1, '2026-01-26 15:11:32.935', '2026-01-26 15:11:32.935', NULL, '13800000000', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user0', NULL, 'male', 2);
INSERT INTO `user` VALUES (2, '2026-01-26 15:11:32.955', '2026-01-26 15:11:32.955', NULL, '13800000001', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user1', NULL, 'male', 1);
INSERT INTO `user` VALUES (3, '2026-01-26 15:11:32.973', '2026-01-26 15:11:32.973', NULL, '13800000002', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user2', NULL, 'male', 1);
INSERT INTO `user` VALUES (4, '2026-01-26 15:11:32.989', '2026-01-26 15:11:32.989', NULL, '13800000003', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user3', NULL, 'male', 1);
INSERT INTO `user` VALUES (5, '2026-01-26 15:11:33.009', '2026-01-26 15:11:33.009', NULL, '13800000004', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user4', NULL, 'male', 1);
INSERT INTO `user` VALUES (6, '2026-01-26 15:11:33.028', '2026-01-26 15:11:33.028', NULL, '13800000005', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user5', NULL, 'male', 1);
INSERT INTO `user` VALUES (7, '2026-01-26 15:11:33.043', '2026-01-26 15:11:33.043', NULL, '13800000006', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user6', NULL, 'male', 1);
INSERT INTO `user` VALUES (8, '2026-01-26 15:11:33.057', '2026-01-26 15:11:33.057', NULL, '13800000007', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user7', NULL, 'male', 1);
INSERT INTO `user` VALUES (9, '2026-01-26 15:11:33.084', '2026-01-26 15:11:33.084', NULL, '13800000008', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user8', NULL, 'male', 1);
INSERT INTO `user` VALUES (10, '2026-01-26 15:11:33.122', '2026-01-26 15:11:33.122', NULL, '13800000009', '$pbkdf2-sha512$Cq3LNk5TGj34v23P$cb7a7bb4f8bc54f8e8281dd9b7ef444de6884b8eeb8d8f51d1317a36647f5ce1', 'user9', NULL, 'male', 1);
INSERT INTO `user` VALUES (11, '2026-01-27 16:46:30.432', '2026-01-27 16:46:30.432', NULL, '13186102266', '$pbkdf2-sha512$s9NWqmKLa122p3zf$50ebff90ed9549cffc1a8fa1879d513d98dcd62700914eef93fdae7d9e1a448a', '13186102266', NULL, 'male', 1);

SET FOREIGN_KEY_CHECKS = 1;
