# 数据库表结构

## 核心表说明

### 1. persons（人员基础表）

- 存储所有人员的基础信息
- `person_type`：1=学生 2=政工 3=维修工
- 通过 students/staff 表关联具体信息和机构

### 2. departments（机构表）

- 使用预排序遍历树（Nested Set）：`tree_left`、`tree_right`、`tree_level`
- `department_type`：0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级
- 支持高效的树查询（查询子树、查询路径）

### 3. students（学生表）

- 关联 person_id
- 单一机构归属：`college_id`、`faculty_id`、`profession_id`、`class_id`

### 4. staff（政工表）

- 关联 person_id
- 单一机构归属：`department_id`、`college_id`、`faculty_id`

### 5. persons_roles（角色表）

- `department_ids`：JSON 格式存储管辖机构 ID，如 `[55,64,56,61,20,46]`
- `person_ids`：JSON 格式存储管辖人员 ID
- 支持灵活的权限配置

### 6. persons_has_roles（用户角色关联表）

- 多对多关系：一个用户可以有多个角色，一个角色可以有多个用户
- 已优化：添加主键和索引

## 数据库优化

### 优化 persons_has_roles 表

```sql
ALTER TABLE `persons_has_roles`
ADD COLUMN `id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
ADD PRIMARY KEY (`id`),
ADD UNIQUE KEY `uk_person_role` (`person_id`, `role_id`),
ADD KEY `idx_person` (`person_id`),
ADD KEY `idx_role` (`role_id`),
ADD COLUMN `created_at` INT(10) UNSIGNED NOT NULL DEFAULT 0,
ADD COLUMN `updated_at` INT(10) UNSIGNED NOT NULL DEFAULT 0;
```

### 优化 persons_roles 表

```sql
ALTER TABLE `persons_roles`
ADD KEY `idx_customer` (`customer_id`),
ADD KEY `idx_parent` (`parent_id`);
```

### 优化 persons 表索引

```sql
ALTER TABLE `persons`
ADD INDEX `idx_customer_type_status` (`customer_id`, `person_type`, `status`),
ADD INDEX `idx_mobile` (`mobile`),
ADD INDEX `idx_email` (`email`);
```

### 优化 students 表索引

```sql
ALTER TABLE `students`
ADD INDEX `idx_class` (`class_id`, `person_id`),
ADD INDEX `idx_college` (`college_id`, `person_id`),
ADD INDEX `idx_profession` (`profession_id`, `person_id`);
```

### 优化 staff 表索引

```sql
ALTER TABLE `staff`
ADD INDEX `idx_department` (`department_id`, `person_id`),
ADD INDEX `idx_college` (`college_id`, `person_id`);
```

## MySQL 主从架构

```
┌──────────┐     同步      ┌──────────┐
│  Master  │ ──────────> │  Slave1  │
│  (写)    │              │  (读)    │
└──────────┘              └──────────┘
                              │
                              │ 同步
                              ▼
                          ┌──────────┐
                          │  Slave2  │
                          │  (读)    │
                          └──────────┘
```

### 读写分离策略

- 写操作：Master
- 读操作：Slave（轮询或权重分配）
- 实时性要求高的读操作：Master
