-- ----------------------------
-- Table structure for banner
-- ----------------------------
DROP TABLE IF EXISTS `banner`;
CREATE TABLE `banner`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `image` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '轮播图图片',
  `url` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '轮播图跳转链接',
  `index` int(11) NOT NULL DEFAULT 0 COMMENT '轮播图顺序',
  `add_time` datetime(3) NULL DEFAULT NULL,
  `is_deleted` int(11) NULL DEFAULT NULL COMMENT '是否删除',
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for brand
-- ----------------------------
DROP TABLE IF EXISTS `brand`;
CREATE TABLE `brand`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '品牌名称',
  `logo` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '品牌logo',
  `add_time` datetime(3) NULL DEFAULT NULL,
  `is_deleted` int(11) NULL DEFAULT NULL COMMENT '是否删除',
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `brand_name`(`name`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for category
-- ----------------------------
DROP TABLE IF EXISTS `category`;
CREATE TABLE `category`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `parent_category_id` int(11) NULL DEFAULT NULL COMMENT '父分类ID',
  `level` int(11) NOT NULL DEFAULT 1 COMMENT '分类级别',
  `is_tab` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否是导航栏',
  `add_time` datetime(3) NULL DEFAULT NULL,
  `is_deleted` int(11) NULL DEFAULT NULL COMMENT '是否删除',
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_category_name`(`name`) USING BTREE,
  INDEX `idx_category_parent_id`(`parent_category_id`) USING BTREE,
  CONSTRAINT `fk_category_parent` FOREIGN KEY (`parent_category_id`) REFERENCES `category` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for good_category_brand
-- ----------------------------
DROP TABLE IF EXISTS `good_category_brand`;
CREATE TABLE `good_category_brand`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category_id` int(11) NOT NULL,
  `brand_id` int(11) NOT NULL,
  `add_time` datetime(3) NULL DEFAULT NULL,
  `is_deleted` int(11) NULL DEFAULT NULL COMMENT '是否删除',
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_category_brand`(`category_id`, `brand_id`) USING BTREE,
  INDEX `idx_gcb_category_id`(`category_id`) USING BTREE,
  INDEX `idx_gcb_brand_id`(`brand_id`) USING BTREE,
  CONSTRAINT `fk_gcb_category` FOREIGN KEY (`category_id`) REFERENCES `category` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `fk_gcb_brand` FOREIGN KEY (`brand_id`) REFERENCES `brand` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for good
-- ----------------------------
DROP TABLE IF EXISTS `good`;
CREATE TABLE `good`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '商品名称',
  `good_sn` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '商品唯一货号, 商家自定义',
  `category_id` int(11) NOT NULL,
  `brand_id` int(11) NOT NULL,
  `on_sale` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否上架',
  `ship_free` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否包邮',
  `is_new` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否新品',
  `is_hot` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否热销',
  `click_num` int(11) NOT NULL DEFAULT 0 COMMENT '点击数',
  `sold_num` int(11) NOT NULL DEFAULT 0 COMMENT '商品销售量',
  `fav_num` int(11) NOT NULL DEFAULT 0 COMMENT '商品收藏数',
  `market_price` float NOT NULL COMMENT '市场价',
  `shop_price` float NOT NULL COMMENT '本店价',
  `good_brief` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '商品简短描述',
  `good_front_image` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '商品封面图',
  `images` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '商品轮播图',
  `desc_images` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '商品详情图',
  `add_time` datetime(3) NULL DEFAULT NULL,
  `is_deleted` int(11) NULL DEFAULT NULL COMMENT '是否删除',
  `update_time` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_good_category_id`(`category_id`) USING BTREE,
  INDEX `idx_good_brand_id`(`brand_id`) USING BTREE,
  CONSTRAINT `fk_good_category` FOREIGN KEY (`category_id`) REFERENCES `category` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_good_brand` FOREIGN KEY (`brand_id`) REFERENCES `brand` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;
