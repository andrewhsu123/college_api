# RoleService - 角色服务

## 职责

- 角色 CRUD
- 角色管辖范围管理
- 角色成员管理

## API 接口

### 角色基础接口

```
GET    /api/roles                           # 角色列表
  参数：customer_id, parent_id, status
GET    /api/roles/{id}                      # 角色详情
POST   /api/roles                           # 创建角色
  Body: {
    customer_id, parent_id, name,
    department_ids: [55, 64, 56],  // JSON格式
    person_ids: [101, 102]         // JSON格式
  }
PUT    /api/roles/{id}                      # 更新角色
DELETE /api/roles/{id}                      # 删除角色（软删除）
```

### 角色管辖范围

```
GET    /api/roles/{id}/scope                # 获取角色管辖范围
  返回：{department_ids: [55,64], person_ids: [101,102]}
PUT    /api/roles/{id}/scope                # 更新角色管辖范围
  Body: {department_ids: [55,64], person_ids: [101,102]}
```

### 角色成员管理

```
GET    /api/roles/{id}/members              # 获取角色下的所有用户
POST   /api/roles/{id}/members              # 批量添加用户到角色
  Body: {person_ids: [123, 124, 125]}
DELETE /api/roles/{id}/members/{person_id}  # 移除角色成员
```

## 数据库表

### persons_roles（角色表）

- `department_ids`：JSON 格式存储管辖机构 ID，如 `[55,64,56,61,20,46]`
- `person_ids`：JSON 格式存储管辖人员 ID
- 支持灵活的权限配置

### persons_has_roles（用户角色关联表）

- 多对多关系：一个用户可以有多个角色，一个角色可以有多个用户
- 已优化：添加主键和索引

## JSON 字段处理

### 存储时转换为 JSON

```php
$role = new PersonsRole();
$role->department_ids = json_encode([55, 64, 56, 61, 20, 46]);
$role->person_ids = json_encode([101, 102, 103]);
$role->save();
```

### 读取时解析 JSON

```php
$role = PersonsRole::find(1);
$departmentIds = json_decode($role->department_ids, true);  // [55, 64, 56, 61, 20, 46]
$personIds = json_decode($role->person_ids, true);          // [101, 102, 103]
```

### MySQL 5.7+ JSON 函数查询

```sql
-- 查询管辖某个机构的所有角色
SELECT id, name, department_ids
FROM persons_roles
WHERE JSON_CONTAINS(department_ids, '55', '$')
  AND customer_id = 1
  AND deleted_at = 0;
```

### 更新 JSON 字段（添加机构）

```sql
UPDATE persons_roles
SET department_ids = JSON_ARRAY_APPEND(department_ids, '$', 99)
WHERE id = 1;
```

### 更新 JSON 字段（移除机构）

```sql
UPDATE persons_roles
SET department_ids = JSON_REMOVE(
  department_ids,
  JSON_UNQUOTE(JSON_SEARCH(department_ids, 'one', 55))
)
WHERE id = 1;
```

## 缓存策略

### 用户角色缓存（Set）

```
person:roles:{person_id} -> [role_id1, role_id2, role_id3]
TTL: 30分钟
示例：person:roles:123 -> [1, 2, 3]
```

### 角色管辖范围缓存（Hash）

```
role:scope:{role_id} -> {
  department_ids: "[55,64,56,61,20,46]",
  person_ids: "[101,102,103]"
}
TTL: 1小时
示例：role:scope:1 -> {department_ids: "[55,64]", person_ids: "[101]"}
```

### 缓存更新

```php
/**
 * 角色分配变更时
 */
function onPersonRoleChanged($personId) {
    // 删除用户角色缓存
    Redis::del("person:roles:{$personId}");
    
    // 删除用户管辖人员缓存
    Redis::del("person:managed:{$personId}");
}

/**
 * 角色管辖范围变更时
 */
function onRoleScopeChanged($roleId) {
    // 删除角色管辖范围缓存
    Redis::del("role:scope:{$roleId}");
    
    // 删除所有拥有该角色的用户的管辖人员缓存
    $personIds = DB::table('persons_has_roles')
        ->where('role_id', $roleId)
        ->pluck('person_id');
    
    foreach ($personIds as $personId) {
        Redis::del("person:managed:{$personId}");
    }
}
```

## 注意事项

- JSON 字段不能建立索引，频繁查询的数据建议缓存
- 如果需要频繁按机构查询角色，考虑使用 `role_has_departments` 关联表
- JSON 字段适合存储不需要频繁查询的配置数据
