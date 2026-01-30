/*
Navicat MySQL Data Transfer

Source Server         : college_dev_base
Source Server Version : 50744
Source Host           : 1.13.252.190:3306
Source Database       : college_dev_base

Target Server Type    : MYSQL
Target Server Version : 50744
File Encoding         : 65001

Date: 2026-01-29 18:00:00
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for personal_access_tokens
-- ----------------------------
DROP TABLE IF EXISTS `personal_access_tokens`;
CREATE TABLE `personal_access_tokens` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `tokenable_type` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型类型',
  `tokenable_id` bigint(20) unsigned NOT NULL COMMENT '模型ID',
  `name` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Token名称',
  `token` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Token哈希值',
  `abilities` text COLLATE utf8mb4_unicode_ci COMMENT '权限',
  `last_used_at` timestamp NULL DEFAULT NULL COMMENT '最后使用时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `personal_access_tokens_token_unique` (`token`),
  KEY `personal_access_tokens_tokenable_type_tokenable_id_index` (`tokenable_type`,`tokenable_id`),
  KEY `personal_access_tokens_expires_at_index` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='个人访问令牌表';
