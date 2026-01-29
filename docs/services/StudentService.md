# StudentService - 学生服务

## 职责

- 学生信息管理
- 学生批量操作

## API 接口

```
GET    /api/students                        # 学生列表
  参数：class_id, profession_id, college_id, grade, enrollment_status
GET    /api/students/{id}                   # 学生详情
POST   /api/students                        # 创建学生
PUT    /api/students/{id}                   # 更新学生
DELETE /api/students/{id}                   # 删除学生
```

### 学生批量操作

```
POST   /api/students/batch-import           # 批量导入学生（Excel）
POST   /api/students/batch-update           # 批量更新学生信息
POST   /api/students/batch-transfer         # 批量转班/转专业
```

## 数据库表

### students（学生表）

- 关联 person_id
- 单一机构归属：`college_id`、`faculty_id`、`profession_id`、`class_id`

### 索引优化

```sql
-- 班级索引（最常用）
ALTER TABLE `students`
ADD INDEX `idx_class_person` (`class_id`, `person_id`);

-- 学院索引
ALTER TABLE `students`
ADD INDEX `idx_college_person` (`college_id`, `person_id`);

-- 专业索引
ALTER TABLE `students`
ADD INDEX `idx_profession_person` (`profession_id`, `person_id`);
```
