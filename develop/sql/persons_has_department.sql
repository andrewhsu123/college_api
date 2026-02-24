/*
Navicat MySQL Data Transfer

Source Server         : localhost_3306
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : college_db_base

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2026-02-01 00:00:00
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for persons_has_department
-- ----------------------------
DROP TABLE IF EXISTS `persons_has_department`;
CREATE TABLE `persons_has_department` (
  `customer_id` int(10) unsigned NOT NULL COMMENT '客户ID(学校ID)',
  `persons_roles_id` int(11) NOT NULL,
  `person_id` int(11) NOT NULL COMMENT '人员ID',
  `department_id` int(11) NOT NULL COMMENT '管辖机构ID'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户管辖机构';
