/*
Navicat MySQL Data Transfer

Source Server         : localhost_3306
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : college_db_base

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2026-01-27 12:57:12
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for persons
-- ----------------------------
DROP TABLE IF EXISTS `persons`;
CREATE TABLE `persons` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '人员ID',
  `customer_id` int(10) unsigned NOT NULL COMMENT '客户ID(学校ID)',
  `person_type` tinyint(4) NOT NULL COMMENT '人员类型:1=学生 2=政工 3=维修工',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '姓名',
  `gender` tinyint(4) DEFAULT NULL COMMENT '性别:1=男,2=女',
  `mobile` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '电子邮箱',
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码',
  `avatar` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态:1=正常,2=禁用',
  `creator_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `created_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_customer_type` (`customer_id`,`person_type`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='人员基础表';
