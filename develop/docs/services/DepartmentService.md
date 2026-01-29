# DepartmentService - 机构服务

## 职责

- 机构 CRUD
- 机构树查询
- 机构人员查询（调用 PersonService）

## API 接口

### 机构基础接口

```
GET    /api/departments/tree                # 获取机构树
  参数：customer_id, department_type（可选）
GET    /api/departments/{id}                # 获取机构详情
GET    /api/departments/{id}/children       # 获取直接子机构
GET    /api/departments/{id}/path           # 获取机构路径（面包屑）
POST   /api/departments                     # 创建机构
PUT    /api/departments/{id}                # 更新机构
DELETE /api/departments/{id}                # 删除机构（软删除）
POST   /api/departments/{id}/move           # 移动机构到新父级
```

### 机构人员查询

```
GET    /api/departments/{id}/persons        # 获取机构下的人员
  参数：include_children（是否包含子机构），person_type, page, page_size
GET    /api/departments/{id}/students       # 获取机构下的学生
GET    /api/departments/{id}/staff          # 获取机构下的政工
```

## 数据库表

### departments（机构表）

- 使用预排序遍历树（Nested Set）：`tree_left`、`tree_right`、`tree_level`
- `department_type`：0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级
- 支持高效的树查询（查询子树、查询路径）

## 预排序遍历树（Nested Set）

### 查询某个机构的所有子机构

```sql
SELECT id, department_name, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_left > (SELECT tree_left FROM departments WHERE id = 55)
  AND tree_right < (SELECT tree_right FROM departments WHERE id = 55)
  AND deleted_at = 0
ORDER BY tree_left;
```

### 查询某个机构的父级路径

```sql
SELECT id, department_name, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_left < (SELECT tree_left FROM departments WHERE id = 55)
  AND tree_right > (SELECT tree_right FROM departments WHERE id = 55)
  AND deleted_at = 0
ORDER BY tree_level;
```

### 查询某个机构的直接子机构

```sql
SELECT id, department_name
FROM departments
WHERE customer_id = 1
  AND parent_id = 55
  AND deleted_at = 0
ORDER BY sort;
```

### 判断机构A是否是机构B的子机构

```sql
SELECT COUNT(*) > 0 AS is_child
FROM departments parent, departments child
WHERE parent.id = 55  -- 机构A
  AND child.id = 64   -- 机构B
  AND child.tree_left > parent.tree_left
  AND child.tree_right < parent.tree_right;
```

## 缓存策略

### 机构树缓存

```php
/**
 * 获取机构树（带缓存）
 */
function getDepartmentTree($customerId, $departmentType = null) {
    $cacheKey = "dept:tree:{$customerId}";
    if ($departmentType !== null) {
        $cacheKey .= ":{$departmentType}";
    }
    
    $cached = Redis::get($cacheKey);
    if ($cached !== null) {
        return json_decode($cached, true);
    }
    
    // 查询所有机构
    $query = DB::table('departments')
        ->where('customer_id', $customerId)
        ->where('deleted_at', 0);
    
    if ($departmentType !== null) {
        $query->where('department_type', $departmentType);
    }
    
    $departments = $query->orderBy('tree_left')->get();
    
    // 构建树结构
    $tree = buildTree($departments);
    
    // 缓存1小时
    Redis::setex($cacheKey, 3600, json_encode($tree));
    
    return $tree;
}
```

### 机构下人员列表缓存（ZSet）

```
dept:persons:{dept_id} -> {person_id: created_at}
TTL: 10分钟
示例：dept:persons:55 -> {123: 1640000000, 124: 1640000001}
```

### 缓存更新

```php
/**
 * 机构变更时
 */
function onDepartmentChanged($customerId) {
    // 删除机构树缓存
    Redis::del("dept:tree:{$customerId}");
    Redis::del("dept:tree:{$customerId}:*");
}

/**
 * 人员机构变更时（学生转班、政工调动）
 */
function onPersonDepartmentChanged($personId, $oldDeptId, $newDeptId) {
    // 删除旧机构的人员列表缓存
    if ($oldDeptId) {
        Redis::del("dept:persons:{$oldDeptId}");
    }
    
    // 删除新机构的人员列表缓存
    if ($newDeptId) {
        Redis::del("dept:persons:{$newDeptId}");
    }
}
```

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 机构树查询 | < 50ms | 1000+ | Redis 1小时 |

## 扩展机构ID列表（包含所有子机构）

```php
/**
 * 扩展机构ID列表（包含所有子机构）
 */
function expandDepartmentIds($deptIds, $customerId) {
    $allDeptIds = [];
    
    foreach ($deptIds as $deptId) {
        // 查询该机构及所有子机构
        $childDeptIds = DB::table('departments')
            ->where('customer_id', $customerId)
            ->where('tree_left', '>=', function($query) use ($deptId) {
                $query->select('tree_left')
                    ->from('departments')
                    ->where('id', $deptId);
            })
            ->where('tree_right', '<=', function($query) use ($deptId) {
                $query->select('tree_right')
                    ->from('departments')
                    ->where('id', $deptId);
            })
            ->where('deleted_at', 0)
            ->pluck('id')
            ->toArray();
        
        $allDeptIds = array_merge($allDeptIds, $childDeptIds);
    }
    
    return array_unique($allDeptIds);
}
```
