-- 校区表
CREATE TABLE `campus_areas` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '区域ID',
  `customer_id` int unsigned NOT NULL COMMENT '客户ID(学校ID)',
  `parent_id` int NOT NULL DEFAULT '0' COMMENT '父级区域ID',
  `area_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '区域名称',
  `level` int NOT NULL DEFAULT '0' COMMENT '0=学校 1=校区 2=片区 3=楼栋 4=楼层 5=宿舍号',
  `created_at` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  PRIMARY KEY (`id`),
  KEY `idx_customer_parent` (`customer_id`,`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='校园区域表';

