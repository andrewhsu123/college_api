-- ============================================
-- 人员中心服务中台 - 数据库优化脚本
-- 执行前请先备份数据库！
-- ============================================

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. persons_has_roles 表优化
-- ============================================
-- 说明：添加主键、索引和时间戳字段，提升查询性能

ALTER TABLE `persons_has_roles`
ADD COLUMN `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT FIRST COMMENT '自增主键',
ADD PRIMARY KEY (`id`),
ADD UNIQUE KEY `uk_person_role` (`person_id`, `role_id`) COMMENT '用户角色唯一索引',
ADD KEY `idx_person` (`person_id`) COMMENT '用户索引',
ADD KEY `idx_role` (`role_id`) COMMENT '角色索引',
ADD COLUMN `created_at` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间' AFTER `role_id`,
ADD COLUMN `updated_at` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间' AFTER `created_at`;

-- 更新表注释
ALTER TABLE `persons_has_roles` COMMENT='用户角色关联表';

-- ============================================
-- 2. persons_roles 表优化
-- ============================================
-- 说明：添加索引，提升按客户ID和父级ID查询的性能

ALTER TABLE `persons_roles`
ADD KEY `idx_customer` (`customer_id`) COMMENT '客户索引',
ADD KEY `idx_parent` (`parent_id`) COMMENT '父级角色索引';

-- ============================================
-- 3. role_has_departments 表优化
-- ============================================
-- 说明：添加主键和索引，提升查询性能

ALTER TABLE `role_has_departments`
ADD COLUMN `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT FIRST COMMENT '自增主键',
ADD PRIMARY KEY (`id`),
ADD UNIQUE KEY `uk_role_dept` (`role_id`, `department_id`) COMMENT '角色机构唯一索引',
ADD KEY `idx_role` (`role_id`) COMMENT '角色索引',
ADD KEY `idx_dept` (`department_id`) COMMENT '机构索引',
ADD COLUMN `created_at` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间' AFTER `department_id`;

-- 更新表注释
ALTER TABLE `role_has_departments` COMMENT='角色管辖机构关联表';

-- ============================================
-- 4. persons 表索引优化
-- ============================================
-- 说明：添加组合索引和单列索引，优化常见查询场景

-- 检查索引是否已存在，避免重复创建
-- 组合索引：客户ID + 人员类型 + 状态（用于列表查询）
ALTER TABLE `persons`
ADD INDEX `idx_customer_type_status` (`customer_id`, `person_type`, `status`) COMMENT '客户类型状态组合索引';

-- 单列索引：手机号（用于登录和搜索）
ALTER TABLE `persons`
ADD INDEX `idx_mobile` (`mobile`) COMMENT '手机号索引';

-- 单列索引：邮箱（用于登录和搜索）
ALTER TABLE `persons`
ADD INDEX `idx_email` (`email`) COMMENT '邮箱索引';

-- ============================================
-- 5. students 表索引优化
-- ============================================
-- 说明：添加机构相关索引，优化按机构查询学生的场景

-- 班级索引（最常用）
ALTER TABLE `students`
ADD INDEX `idx_class_person` (`class_id`, `person_id`) COMMENT '班级人员索引';

-- 学院索引
ALTER TABLE `students`
ADD INDEX `idx_college_person` (`college_id`, `person_id`) COMMENT '学院人员索引';

-- 专业索引
ALTER TABLE `students`
ADD INDEX `idx_profession_person` (`profession_id`, `person_id`) COMMENT '专业人员索引';

-- 学号索引（如果不存在）
-- ALTER TABLE `students`
-- ADD INDEX `idx_student_no` (`student_no`) COMMENT '学号索引';

-- ============================================
-- 6. staff 表索引优化
-- ============================================
-- 说明：添加机构相关索引，优化按机构查询政工的场景

-- 部门索引（最常用）
ALTER TABLE `staff`
ADD INDEX `idx_dept_person` (`department_id`, `person_id`) COMMENT '部门人员索引';

-- 学院索引
ALTER TABLE `staff`
ADD INDEX `idx_college_person` (`college_id`, `person_id`) COMMENT '学院人员索引';

-- 工号索引（如果不存在）
-- ALTER TABLE `staff`
-- ADD INDEX `idx_staff_no` (`staff_no`) COMMENT '工号索引';

-- ============================================
-- 7. departments 表索引优化（可选）
-- ============================================
-- 说明：如果现有索引不够，可以添加以下索引

-- 树查询组合索引
-- ALTER TABLE `departments`
-- ADD INDEX `idx_tree_query` (`customer_id`, `tree_left`, `tree_right`) COMMENT '树查询索引';

-- 父级索引（如果不存在）
-- ALTER TABLE `departments`
-- ADD INDEX `idx_parent` (`parent_id`) COMMENT '父级索引';

-- ============================================
-- 8. 验证索引创建结果
-- ============================================

-- 查看 persons_has_roles 表结构
SHOW CREATE TABLE `persons_has_roles`;

-- 查看 persons_roles 表索引
SHOW INDEX FROM `persons_roles`;

-- 查看 role_has_departments 表结构
SHOW CREATE TABLE `role_has_departments`;

-- 查看 persons 表索引
SHOW INDEX FROM `persons`;

-- 查看 students 表索引
SHOW INDEX FROM `students`;

-- 查看 staff 表索引
SHOW INDEX FROM `staff`;

-- ============================================
-- 9. 性能分析建议
-- ============================================

-- 分析表（更新索引统计信息）
ANALYZE TABLE `persons_has_roles`;
ANALYZE TABLE `persons_roles`;
ANALYZE TABLE `role_has_departments`;
ANALYZE TABLE `persons`;
ANALYZE TABLE `students`;
ANALYZE TABLE `staff`;
ANALYZE TABLE `departments`;

-- ============================================
-- 完成
-- ============================================

SET FOREIGN_KEY_CHECKS = 1;

-- 优化完成！
-- 建议：
-- 1. 执行后观察慢查询日志，进一步优化
-- 2. 定期执行 ANALYZE TABLE 更新统计信息
-- 3. 监控索引使用情况，删除未使用的索引
