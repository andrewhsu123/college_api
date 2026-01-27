/*
Navicat MySQL Data Transfer

Source Server         : localhost_3306
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : college_db_base

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2026-01-27 12:55:37
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for departments
-- ----------------------------
DROP TABLE IF EXISTS `departments`;
CREATE TABLE `departments` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `customer_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '客户ID(学校ID\\tree_group)',
  `parent_id` int(11) NOT NULL DEFAULT '0' COMMENT '父级ID',
  `tree_level` int(10) unsigned DEFAULT NULL COMMENT '等级',
  `tree_left` int(10) unsigned DEFAULT NULL COMMENT '左值（预排序遍历树）',
  `tree_right` int(10) unsigned DEFAULT NULL COMMENT '右值（预排序遍历树）',
  `recommend_num` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '真实的父级机构数',
  `department_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
  `department_type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '机构类型:0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级',
  `principal` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人',
  `mobile` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人联系方式',
  `email` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱',
  `status` smallint(6) NOT NULL DEFAULT '1' COMMENT '1 正常 2 停用',
  `sort` int(11) NOT NULL DEFAULT '1' COMMENT '排序',
  `created_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  PRIMARY KEY (`id`),
  KEY `idx_customer` (`customer_id`),
  KEY `idx_parent` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门/机构表';
