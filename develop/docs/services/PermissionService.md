# PermissionService - 权限服务

## 职责

- 权限计算（用户 → 角色 → 管辖范围）
- 权限校验
- 管辖人员/机构查询

## API 接口

### 权限查询

```
GET    /api/permissions/my-roles            # 获取当前用户的所有角色
GET    /api/permissions/my-scope            # 获取当前用户的管辖范围
  返回：{department_ids: [...], person_ids: [...]}
GET    /api/permissions/my-departments      # 获取当前用户可管辖的机构列表
GET    /api/permissions/my-persons          # 获取当前用户可管辖的人员列表
  参数：person_type, page, page_size
```

### 权限校验

```
POST   /api/permissions/check               # 检查权限
  Body: {
    action: "view_person",  // 操作类型
    target_id: 123          // 目标ID（人员ID或机构ID）
  }
  返回：{allowed: true/false, reason: "..."}
```

### 批量权限校验

```
POST   /api/permissions/batch-check         # 批量检查权限
  Body: {
    action: "view_person",
    target_ids: [123, 124, 125]
  }
  返回：{
    allowed_ids: [123, 124],
    denied_ids: [125]
  }
```

## 权限计算流程

```
用户查询管辖人员
    ↓
1. 从缓存获取用户角色（person:roles:{person_id}）
   缓存未命中 → 查询 persons_has_roles 表
    ↓
2. 从缓存获取角色管辖范围（role:scope:{role_id}）
   缓存未命中 → 查询 persons_roles 表
    ↓
3. 解析 JSON 字段（department_ids、person_ids）
    ↓
4. 根据机构 ID 查询人员
   - 从缓存获取机构下人员（dept:persons:{dept_id}）
   - 缓存未命中 → 查询 students/staff 表
    ↓
5. 合并直接管辖的人员 ID（person_ids）
    ↓
6. 去重并返回人员列表
    ↓
7. 缓存最终结果（person:managed:{person_id}）
```

## 核心实现

### 获取用户管辖的所有人员ID

```php
/**
 * 获取用户管辖的所有人员ID
 * @param int $personId 用户ID
 * @param int $customerId 客户ID
 * @return array 人员ID列表
 */
function getManagedPersonIds($personId, $customerId) {
    // 1. 尝试从缓存获取最终结果
    $cacheKey = "person:managed:{$personId}";
    $cachedResult = Redis::get($cacheKey);
    if ($cachedResult !== null) {
        return json_decode($cachedResult, true);
    }
    
    // 2. 获取用户的所有角色
    $roleIds = getUserRoles($personId);
    if (empty($roleIds)) {
        return [];
    }
    
    // 3. 获取角色管辖范围
    $managedPersonIds = [];
    $managedDeptIds = [];
    
    foreach ($roleIds as $roleId) {
        $scope = getRoleScope($roleId, $customerId);
        if (!empty($scope['person_ids'])) {
            $managedPersonIds = array_merge($managedPersonIds, $scope['person_ids']);
        }
        if (!empty($scope['department_ids'])) {
            $managedDeptIds = array_merge($managedDeptIds, $scope['department_ids']);
        }
    }
    
    // 4. 根据机构查询人员
    if (!empty($managedDeptIds)) {
        $deptPersonIds = getPersonIdsByDepartments($managedDeptIds, $customerId);
        $managedPersonIds = array_merge($managedPersonIds, $deptPersonIds);
    }
    
    // 5. 去重
    $managedPersonIds = array_unique($managedPersonIds);
    
    // 6. 缓存结果（5分钟）
    Redis::setex($cacheKey, 300, json_encode($managedPersonIds));
    
    return $managedPersonIds;
}
```

### 获取用户的所有角色

```php
function getUserRoles($personId) {
    $cacheKey = "person:roles:{$personId}";
    $cached = Redis::smembers($cacheKey);
    if (!empty($cached)) {
        return $cached;
    }
    
    // 从数据库查询
    $roles = DB::table('persons_has_roles')
        ->where('person_id', $personId)
        ->pluck('role_id')
        ->toArray();
    
    // 缓存30分钟
    if (!empty($roles)) {
        Redis::sadd($cacheKey, ...$roles);
        Redis::expire($cacheKey, 1800);
    }
    
    return $roles;
}
```

### 获取角色管辖范围

```php
function getRoleScope($roleId, $customerId) {
    $cacheKey = "role:scope:{$roleId}";
    $cached = Redis::hgetall($cacheKey);
    if (!empty($cached)) {
        return [
            'department_ids' => json_decode($cached['department_ids'] ?? '[]', true),
            'person_ids' => json_decode($cached['person_ids'] ?? '[]', true),
        ];
    }
    
    // 从数据库查询
    $role = DB::table('persons_roles')
        ->where('id', $roleId)
        ->where('customer_id', $customerId)
        ->where('deleted_at', 0)
        ->first();
    
    if (!$role) {
        return ['department_ids' => [], 'person_ids' => []];
    }
    
    $scope = [
        'department_ids' => json_decode($role->department_ids ?? '[]', true),
        'person_ids' => json_decode($role->person_ids ?? '[]', true),
    ];
    
    // 缓存1小时
    Redis::hmset($cacheKey, [
        'department_ids' => $role->department_ids ?? '[]',
        'person_ids' => $role->person_ids ?? '[]',
    ]);
    Redis::expire($cacheKey, 3600);
    
    return $scope;
}
```

### 根据机构ID列表查询人员ID

```php
function getPersonIdsByDepartments($deptIds, $customerId) {
    $personIds = [];
    
    foreach ($deptIds as $deptId) {
        // 尝试从缓存获取
        $cacheKey = "dept:persons:{$deptId}";
        $cached = Redis::zrange($cacheKey, 0, -1);
        
        if (!empty($cached)) {
            $personIds = array_merge($personIds, $cached);
            continue;
        }
        
        // 从数据库查询学生
        $studentIds = DB::table('students')
            ->join('persons', 'students.person_id', '=', 'persons.id')
            ->where('students.class_id', $deptId)
            ->where('persons.status', 1)
            ->where('persons.deleted_at', 0)
            ->pluck('students.person_id')
            ->toArray();
        
        // 从数据库查询政工
        $staffIds = DB::table('staff')
            ->join('persons', 'staff.person_id', '=', 'persons.id')
            ->where('staff.department_id', $deptId)
            ->where('persons.status', 1)
            ->where('persons.deleted_at', 0)
            ->pluck('staff.person_id')
            ->toArray();
        
        $deptPersonIds = array_merge($studentIds, $staffIds);
        
        // 缓存10分钟
        if (!empty($deptPersonIds)) {
            $now = time();
            $zsetData = [];
            foreach ($deptPersonIds as $pid) {
                $zsetData[$pid] = $now;
            }
            Redis::zadd($cacheKey, ...$zsetData);
            Redis::expire($cacheKey, 600);
        }
        
        $personIds = array_merge($personIds, $deptPersonIds);
    }
    
    return $personIds;
}
```

### 检查用户是否有权限查看目标人员

```php
function canViewPerson($currentPersonId, $targetPersonId, $customerId) {
    $managedPersonIds = getManagedPersonIds($currentPersonId, $customerId);
    return in_array($targetPersonId, $managedPersonIds);
}
```

## 缓存策略

### 用户管辖人员缓存（Set）

```
person:managed:{person_id} -> [managed_person_id1, managed_person_id2...]
TTL: 5分钟
示例：person:managed:123 -> [201, 202, 203, 204]
```

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 管辖人员查询 | < 100ms | 500+ | Redis 5分钟 |
| 权限校验 | < 20ms | 2000+ | Redis 30分钟 |

## SQL 查询示例

### 查询用户管辖的人员

```sql
-- 步骤1：查询用户的所有角色
SELECT role_id 
FROM persons_has_roles 
WHERE person_id = 123;

-- 步骤2：查询角色管辖的机构和人员
SELECT id, department_ids, person_ids 
FROM persons_roles 
WHERE id IN (1, 2, 3)
  AND customer_id = 1
  AND deleted_at = 0;

-- 步骤3：查询机构下的学生
SELECT DISTINCT s.person_id, p.name, p.mobile, s.student_no, s.class_id
FROM students s
INNER JOIN persons p ON s.person_id = p.id
WHERE s.class_id IN (55, 64, 56, 61, 20, 46)
  AND p.status = 1
  AND p.deleted_at = 0;

-- 步骤4：查询机构下的政工
SELECT DISTINCT st.person_id, p.name, p.mobile, st.staff_no, st.department_id
FROM staff st
INNER JOIN persons p ON st.person_id = p.id
WHERE st.department_id IN (55, 64, 56, 61, 20, 46)
  AND p.status = 1
  AND p.deleted_at = 0;

-- 步骤5：直接管辖的人员
SELECT id, name, mobile, person_type
FROM persons
WHERE id IN (101, 102, 103)
  AND status = 1
  AND deleted_at = 0;
```
