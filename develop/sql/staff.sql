/*
Navicat MySQL Data Transfer

Source Server         : localhost_3306
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : college_db_base

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2026-01-27 12:57:18
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for staff
-- ----------------------------
DROP TABLE IF EXISTS `staff`;
CREATE TABLE `staff` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `university_id` int unsigned NOT NULL COMMENT '高校ID(关联customers表)',
  `person_id` int unsigned NOT NULL COMMENT '关联人员ID',
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名(冗余字段)',
  `staff_no` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工号',
  `department_id` int unsigned DEFAULT NULL COMMENT '所属部门ID(关联departments表)',
  `college_id` int unsigned DEFAULT NULL COMMENT '所属学院ID(关联departments表)',
  `faculty_id` int unsigned DEFAULT NULL COMMENT '所属系ID(关联departments表)',
  `created_at` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_college` (`university_id`,`college_id`) USING BTREE,
  KEY `idx_department` (`university_id`,`department_id`) USING BTREE,
  KEY `idx_faculty` (`university_id`,`faculty_id`),
  KEY `idx_person` (`university_id`,`person_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='政工表';

