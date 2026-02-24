/*
Navicat MySQL Data Transfer

Source Server         : localhost_3306
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : college_db_base

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2026-01-27 12:57:25
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for students
-- ----------------------------
DROP TABLE IF EXISTS `students`;
CREATE TABLE `students` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `university_id` int(10) unsigned NOT NULL COMMENT '高校ID(关联customers表)',
  `person_id` int(10) unsigned NOT NULL COMMENT '关联人员ID',
  `area_id` int(11) DEFAULT NULL COMMENT '校区ID(关联campus_areas表)',
  `name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '姓名(冗余字段)',
  `student_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '学号',
  `grade` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '年级',
  `education_level` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '教育层次',
  `school_system` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '学制',
  `id_card` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '身份证号',
  `admission_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '录取编号',
  `exam_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '准考证号',
  `enrollment_status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '学籍状态:1=在读,2=休学,3=毕业,4=退学,5=停学,6=复学,7=未报到,8=结业,9=肄业,10=转学,11=死亡,12=开除,13=参军,14=保留学籍,15=其他',
  `is_enrolled` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=已报到 2=未报到',
  `college_id` int(10) unsigned NOT NULL COMMENT '学院ID(关联departments表)',
  `faculty_id` int(10) unsigned DEFAULT NULL COMMENT '系ID(关联departments表)',
  `profession_id` int(10) unsigned NOT NULL COMMENT '专业ID(关联departments表)',
  `class_id` int(10) unsigned DEFAULT NULL COMMENT '班级ID(关联departments表)',
  `created_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_student_no` (`student_no`) USING BTREE,
  KEY `idx_person` (`person_id`) USING BTREE,
  KEY `idx_org_path` (`university_id`,`college_id`,`faculty_id`,`profession_id`) USING BTREE,
  KEY `idx_class` (`class_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='学生表';
