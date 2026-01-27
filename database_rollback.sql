-- ============================================
-- 人员中心服务中台 - 数据库优化回滚脚本
-- 如果优化后出现问题，可以使用此脚本回滚
-- ============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. 回滚 persons_has_roles 表
-- ============================================

-- 删除新增的索引
ALTER TABLE `persons_has_roles`
DROP INDEX `uk_person_role`,
DROP INDEX `idx_person`,
DROP INDEX `idx_role`;

-- 删除主键（需要先删除 AUTO_INCREMENT）
ALTER TABLE `persons_has_roles`
MODIFY COLUMN `id` INT(10) UNSIGNED NOT NULL;

ALTER TABLE `persons_has_roles`
DROP PRIMARY KEY;

-- 删除新增的列
ALTER TABLE `persons_has_roles`
DROP COLUMN `id`,
DROP COLUMN `created_at`,
DROP COLUMN `updated_at`;

-- ============================================
-- 2. 回滚 persons_roles 表
-- ============================================

ALTER TABLE `persons_roles`
DROP INDEX `idx_customer`,
DROP INDEX `idx_parent`;

-- ============================================
-- 3. 回滚 role_has_departments 表
-- ============================================

-- 删除索引
ALTER TABLE `role_has_departments`
DROP INDEX `uk_role_dept`,
DROP INDEX `idx_role`,
DROP INDEX `idx_dept`;

-- 删除主键
ALTER TABLE `role_has_departments`
MODIFY COLUMN `id` INT(10) UNSIGNED NOT NULL;

ALTER TABLE `role_has_departments`
DROP PRIMARY KEY;

-- 删除新增的列
ALTER TABLE `role_has_departments`
DROP COLUMN `id`,
DROP COLUMN `created_at`;

-- ============================================
-- 4. 回滚 persons 表
-- ============================================

ALTER TABLE `persons`
DROP INDEX `idx_customer_type_status`,
DROP INDEX `idx_mobile`,
DROP INDEX `idx_email`;

-- ============================================
-- 5. 回滚 students 表
-- ============================================

ALTER TABLE `students`
DROP INDEX `idx_class_person`,
DROP INDEX `idx_college_person`,
DROP INDEX `idx_profession_person`;

-- ============================================
-- 6. 回滚 staff 表
-- ============================================

ALTER TABLE `staff`
DROP INDEX `idx_dept_person`,
DROP INDEX `idx_college_person`;

-- ============================================
-- 验证回滚结果
-- ============================================

SHOW CREATE TABLE `persons_has_roles`;
SHOW CREATE TABLE `persons_roles`;
SHOW CREATE TABLE `role_has_departments`;
SHOW INDEX FROM `persons`;
SHOW INDEX FROM `students`;
SHOW INDEX FROM `staff`;

SET FOREIGN_KEY_CHECKS = 1;

-- 回滚完成！
