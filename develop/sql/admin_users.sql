/*
Navicat MySQL Data Transfer

Source Server         : college_dev_base
Source Server Version : 50744
Source Host           : 1.13.252.190:3306
Source Database       : college_dev_base

Target Server Type    : MYSQL
Target Server Version : 50744
File Encoding         : 65001

Date: 2026-01-29 17:50:01
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for admin_users
-- ----------------------------
DROP TABLE IF EXISTS `admin_users`;
CREATE TABLE `admin_users` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '昵称',
  `password` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码',
  `email` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '邮箱',
  `mobile` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `wx_pc_openid` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '微信网页授权的 openid',
  `unionid` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '微信唯一用户标识 unionid',
  `avatar` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像',
  `remember_token` varchar(1000) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'token',
  `login_token` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '免密登录token',
  `department_id` int(11) NOT NULL DEFAULT '0' COMMENT '部门ID',
  `customer_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '客户ID(学校ID)',
  `creator_id` int(11) NOT NULL DEFAULT '0',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态:1=正常,2=禁用',
  `login_ip` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '登录IP',
  `login_at` int(11) NOT NULL DEFAULT '0' COMMENT '登录时间',
  `created_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  PRIMARY KEY (`id`),
  KEY `idx_customer` (`customer_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
