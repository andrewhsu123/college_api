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

### 5. persons_roles（管辖角色表）

- `customer_id`：客户ID（学校ID）
- `parent_id`：上级角色组ID
- `name`：角色名称
- `permissions`：菜单权限，逗号分隔的菜单ID（如 "1,2,3,4"）
- **注意**：已删除 `department_ids` 和 `person_ids` 字段

### 6. persons_has_roles（用户角色关联表）

- `customer_id`：客户ID（学校ID）- 新增字段
- `person_id`：人员ID
- `role_id`：角色ID
- 查询时先查 `customer_id`，再查 `role_id` 以命中索引

### 7. persons_has_department（人员角色管辖机构关系表）

- `customer_id`：客户ID（学校ID）- 新增字段
- `persons_roles_id`：角色ID - 新增字段
- `person_id`：人员ID
- `department_id`：管辖机构ID
- **新逻辑**：角色 + 人员 决定管辖机构关系

## 管辖权限新逻辑

### 权限查询流程

1. 通过 `persons_has_roles` 表查询人员拥有的角色（先查 `customer_id`，再查 `person_id`）
2. 通过 `persons_roles` 表获取角色详情（名称、上级角色组、菜单权限）
3. 通过 `persons_has_department` 表查询角色+人员对应的管辖机构

### 权限输出格式

```json
{
  "managed_roles": [
    {
      "id": 1,
      "parent_id": 0,
      "parent_name": "",
      "name": "辅导员",
      "departments": [
        {
          "id": 55,
          "parent_id": 10,
          "department_name": "计算机学院",
          "department_type": 2,
          "status": 1
        }
      ]
    }
  ],
  "managed_menu": [1, 2, 3, 4]
}
```

## 数据库优化

### 优化 persons_has_roles 表

```sql
-- 新表结构
CREATE TABLE `persons_has_roles` (
  `customer_id` int(10) unsigned NOT NULL COMMENT '客户ID(学校ID)',
  `person_id` int(11) NOT NULL COMMENT '人员ID',
  `role_id` int(11) NOT NULL COMMENT '角色ID',
  KEY `idx_customer_person` (`customer_id`, `person_id`),
  KEY `idx_customer_role` (`customer_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色';
```

### 优化 persons_has_department 表

```sql
-- 新表结构
CREATE TABLE `persons_has_department` (
  `customer_id` int(10) unsigned NOT NULL COMMENT '客户ID(学校ID)',
  `persons_roles_id` int(11) NOT NULL COMMENT '角色ID',
  `person_id` int(11) NOT NULL COMMENT '人员ID',
  `department_id` int(11) NOT NULL COMMENT '管辖机构ID',
  KEY `idx_customer_person_role` (`customer_id`, `person_id`, `persons_roles_id`),
  KEY `idx_customer_department` (`customer_id`, `department_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户管辖机构';
```

### 优化 persons_roles 表

```sql
-- 新表结构
CREATE TABLE `persons_roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '管辖角色Id',
  `customer_id` int(11) NOT NULL,
  `parent_id` int(11) NOT NULL DEFAULT '0' COMMENT '上级角色组',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `permissions` text COLLATE utf8mb4_unicode_ci COMMENT '角色权限(菜单ID，逗号分隔)',
  `created_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '软删除',
  PRIMARY KEY (`id`),
  KEY `idx_customer` (`customer_id`),
  KEY `idx_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管辖角色';
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
