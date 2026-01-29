# 缓存策略

## Redis 缓存设计

### 1. 用户角色缓存（Set）

```
person:roles:{person_id} -> [role_id1, role_id2, role_id3]
TTL: 30分钟
示例：person:roles:123 -> [1, 2, 3]
```

### 2. 角色管辖范围缓存（Hash）

```
role:scope:{role_id} -> {
  department_ids: "[55,64,56,61,20,46]",
  person_ids: "[101,102,103]"
}
TTL: 1小时
示例：role:scope:1 -> {department_ids: "[55,64]", person_ids: "[101]"}
```

### 3. 机构下人员列表缓存（ZSet，按创建时间排序）

```
dept:persons:{dept_id} -> {person_id: created_at}
TTL: 10分钟
示例：dept:persons:55 -> {123: 1640000000, 124: 1640000001}
```

### 4. 人员基础信息缓存（Hash）

```
person:info:{person_id} -> {id, name, mobile, avatar, person_type, status}
TTL: 1小时
示例：person:info:123 -> {id:123, name:"张三", mobile:"13800138000"}
```

### 5. 用户管辖人员缓存（Set）- 最终结果缓存

```
person:managed:{person_id} -> [managed_person_id1, managed_person_id2...]
TTL: 5分钟
示例：person:managed:123 -> [201, 202, 203, 204]
```

## 缓存更新策略

### 缓存失效触发点

#### 人员信息变更时

```php
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

#### 角色分配变更时

```php
function onPersonRoleChanged($personId) {
    // 删除用户角色缓存
    Redis::del("person:roles:{$personId}");
    
    // 删除用户管辖人员缓存
    Redis::del("person:managed:{$personId}");
}
```

#### 角色管辖范围变更时

```php
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

#### 人员机构变更时（学生转班、政工调动）

```php
function onPersonDepartmentChanged($personId, $oldDeptId, $newDeptId) {
    // 删除旧机构的人员列表缓存
    if ($oldDeptId) {
        Redis::del("dept:persons:{$oldDeptId}");
    }
    
    // 删除新机构的人员列表缓存
    if ($newDeptId) {
        Redis::del("dept:persons:{$newDeptId}");
    }
    
    // 删除相关管理员的管辖人员缓存
    $managerIds = getManagersByDepartments([$oldDeptId, $newDeptId]);
    foreach ($managerIds as $managerId) {
        Redis::del("person:managed:{$managerId}");
    }
}
```

#### 机构变更时

```php
function onDepartmentChanged($customerId) {
    // 删除机构树缓存
    Redis::del("dept:tree:{$customerId}");
    Redis::del("dept:tree:{$customerId}:*");
}
```

## 延迟双删策略

保证缓存一致性：

```php
function updatePersonInfo($personId, $data) {
    // 1. 删除缓存
    Redis::del("person:info:{$personId}");
    
    // 2. 更新数据库
    DB::table('persons')->where('id', $personId)->update($data);
    
    // 3. 延迟500ms再次删除缓存（防止并发读取到旧数据）
    sleep(0.5);
    Redis::del("person:info:{$personId}");
}
```

## Redis 集群部署

```
┌─────────┐  ┌─────────┐  ┌─────────┐
│ Master1 │  │ Master2 │  │ Master3 │
│ Slave1  │  │ Slave2  │  │ Slave3  │
└─────────┘  └─────────┘  └─────────┘
```

- 使用 Redis Cluster 或 Sentinel
- 主从复制 + 自动故障转移
- 分片存储，提高容量和性能

## 缓存优化技巧

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

### 缓存预热

系统启动时加载热点数据，提高缓存命中率。

### 缓存穿透防护

使用布隆过滤器防止缓存穿透攻击。
