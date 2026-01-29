/*
Navicat MySQL Data Transfer

Source Server         : localhost_3306
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : college_db_base

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2026-01-27 19:38:36
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for role_has_departments
-- ----------------------------
DROP TABLE IF EXISTS `role_has_departments`;
CREATE TABLE `role_has_departments` (
  `role_id` int(11) NOT NULL COMMENT 'roles primary key',
  `department_id` int(11) NOT NULL COMMENT 'departments primary key'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='role relate departments';
