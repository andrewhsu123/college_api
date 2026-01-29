# PersonService - 人员服务

## 职责

- 人员 CRUD
- 人员搜索（调用 SearchService）
- 人员角色管理（调用 RoleService）
- 学生/政工管理

## API 接口

### 人员基础接口

```
GET    /api/persons/{id}                    # 获取人员详情
GET    /api/persons                         # 人员列表（支持分页、筛选）
POST   /api/persons                         # 创建人员
PUT    /api/persons/{id}                    # 更新人员
DELETE /api/persons/{id}                    # 删除人员（软删除）
```

### 人员搜索

```
GET    /api/persons/search                  # 搜索人员（ES，支持姓名/手机/学号/工号）
  参数：keyword, person_type, status, page, page_size
```

### 人员角色管理

```
GET    /api/persons/{id}/roles              # 获取人员的所有角色
POST   /api/persons/{id}/roles              # 为人员分配角色
  Body: {role_ids: [1, 2, 3]}
DELETE /api/persons/{id}/roles/{role_id}    # 移除人员角色
```

### 人员机构查询

```
GET    /api/persons/{id}/departments        # 获取人员所属机构
  返回：学生返回班级/专业/系/学院，政工返回部门/学院
```

## 数据库表

### persons（人员基础表）

- 存储所有人员的基础信息
- `person_type`：1=学生 2=政工 3=维修工
- 通过 students/staff 表关联具体信息和机构

### 索引优化

```sql
-- 组合索引：客户ID + 人员类型 + 状态（用于列表查询）
ALTER TABLE `persons`
ADD INDEX `idx_customer_type_status` (`customer_id`, `person_type`, `status`);

-- 单列索引：手机号（用于登录和搜索）
ALTER TABLE `persons`
ADD INDEX `idx_mobile` (`mobile`);

-- 单列索引：邮箱（用于登录和搜索）
ALTER TABLE `persons`
ADD INDEX `idx_email` (`email`);
```

## 缓存策略

### 人员基础信息缓存（Hash）

```
person:info:{person_id} -> {id, name, mobile, avatar, person_type, status}
TTL: 1小时
示例：person:info:123 -> {id:123, name:"张三", mobile:"13800138000"}
```

### 缓存更新

```php
/**
 * 人员信息变更时
 */
function onPersonUpdated($personId) {
    // 删除人员信息缓存
    Redis::del("person:info:{$personId}");
    
    // 删除相关管理员的管辖人员缓存
    $managerIds = getManagersByPerson($personId);
    foreach ($managerIds as $managerId) {
        Redis::del("person:managed:{$managerId}");
    }
}
```

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 人员详情查询 | < 50ms | 1000+ | Redis 1小时 |
| 人员列表查询 | < 100ms | 500+ | 数据库查询 |

## 实现要点

### 批量查询优化

```php
// 不好的做法：N+1查询
$persons = Person::where('customer_id', 1)->get();
foreach ($persons as $person) {
    $person->student = Student::where('person_id', $person->id)->first();
}

// 好的做法：预加载
$persons = Person::with('student')->where('customer_id', 1)->get();
```

### 使用 Redis Pipeline 批量操作

```php
// 批量获取人员信息
$personIds = [123, 124, 125, 126];
$pipeline = Redis::pipeline();
foreach ($personIds as $personId) {
    $pipeline->hgetall("person:info:{$personId}");
}
$results = $pipeline->execute();
```
