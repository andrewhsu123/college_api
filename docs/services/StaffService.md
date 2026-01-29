# StaffService - 政工服务

## 职责

- 政工信息管理
- 政工批量操作

## API 接口

```
GET    /api/staff                           # 政工列表
  参数：department_id, college_id, faculty_id
GET    /api/staff/{id}                      # 政工详情
POST   /api/staff                           # 创建政工
PUT    /api/staff/{id}                      # 更新政工
DELETE /api/staff/{id}                      # 删除政工
```

### 政工批量操作

```
POST   /api/staff/batch-import              # 批量导入政工（Excel）
POST   /api/staff/batch-transfer            # 批量调动部门
```

## 数据库表

### staff（政工表）

- 关联 person_id
- 单一机构归属：`department_id`、`college_id`、`faculty_id`

### 索引优化

```sql
-- 部门索引（最常用）
ALTER TABLE `staff`
ADD INDEX `idx_dept_person` (`department_id`, `person_id`);

-- 学院索引
ALTER TABLE `staff`
ADD INDEX `idx_college_person` (`college_id`, `person_id`);
```
