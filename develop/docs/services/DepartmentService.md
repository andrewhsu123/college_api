# DepartmentService - 机构服务

## 职责

- 机构树查询（学校 -> 行政机构/组织机构的两级树形结构）
- 机构搜索（名称模糊查询、类型查询）
- 支持学校管理员和政工人员两种角色的权限过滤

## API 接口

### 学校管理员接口（/api/base）

```
GET    /api/base/departments/tree           # 获取机构树（两级结构）
  认证：Bearer Token (学校管理员)
  参数：customer_id
  返回：树形结构数据
  说明：第一级为学校，第二级为行政机构和组织机构
  权限：查看所有机构
  
GET    /api/base/departments/list           # 搜索机构列表
  认证：Bearer Token (学校管理员)
  参数：customer_id, keyword（机构名称模糊查询，可选）, department_type（机构类型，可选）
  返回：机构列表数组
  说明：支持名称模糊查询和类型筛选
  权限：查看所有机构
```

### 政工人员接口（/api/staff）

```
GET    /api/staff/departments/tree          # 获取机构树（两级结构）
  认证：Bearer Token (政工人员)
  参数：customer_id
  返回：树形结构数据
  说明：第一级为学校，第二级为行政机构和组织机构
  权限：仅查看被授权的机构及其子机构
  
GET    /api/staff/departments/list          # 搜索机构列表
  认证：Bearer Token (政工人员)
  参数：customer_id, keyword（机构名称模糊查询，可选）, department_type（机构类型，可选）
  返回：机构列表数组
  说明：支持名称模糊查询和类型筛选
  权限：仅查看被授权的机构及其子机构
```

## 用户角色说明

### 学校管理员
- 登录接口：`GET /api/base/info`
- 认证方式：`Authorization: Bearer {token}`
- 权限范围：可以查看所有机构
- 用户标识：通过 `admin_users` 表的 `user_id` 获取 `customer_id`

```sql
-- 获取学校管理员的 customer_id
SELECT customer_id FROM admin_users WHERE id = {user_id};
```

### 政工人员
- 登录接口：`GET /api/staff/info`
- 认证方式：`Authorization: Bearer {token}`
- 权限范围：只能查看被授权的机构及其子机构
- 用户标识：通过 `persons` 表的 `person_id` 获取 `customer_id` 和权限

```sql
-- 获取政工人员的 customer_id
SELECT customer_id FROM persons WHERE id = {person_id};

-- 获取政工人员可见的部门ID列表（两个来源）
-- 来源1：通过角色关联的部门权限
SELECT department_ids 
FROM persons_roles 
WHERE customer_id = {customer_id} 
  AND id IN (
    SELECT role_id 
    FROM persons_has_roles 
    WHERE person_id = {person_id}
  );

-- 来源2：直接分配给人员的部门权限
SELECT department_id 
FROM persons_has_department 
WHERE person_id = {person_id};
```

**查询结果示例：**
```json
// 角色关联的部门权限
[
  [55, 64, 56, 61, 20, 46],
  [55]
]

// 直接分配的部门权限
[10, 20, 30]
```

需要将两个来源的部门ID合并去重，得到政工人员最终可见的机构ID列表：`[55, 64, 56, 61, 20, 46, 10, 30]`

## 数据库表

### departments（机构表）

- 使用预排序遍历树（Nested Set）：`tree_left`、`tree_right`、`tree_level`
- `department_type`：0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级
- `recommend_num`：下级机构数量（0表示没有下级）
- 支持高效的树查询（查询子树、查询路径）

## 机构树结构说明

### 树形结构规则

机构树采用两级展示结构：

1. **第一级：学校** (`tree_level = 1`, `department_type = 0`)
2. **第二级：行政机构和组织机构** (`tree_level = 3`)
   - 行政机构：`department_type = 1`
   - 组织机构：`department_type != 1`（学院、系、专业、班级）

### 数据组织规则

- 所有 `tree_level = 3` 的机构都作为学校的直接子节点展示，并且 `parent_id` 改为指向学校的 id
- `tree_level > 3` 的机构按照实际的 `parent_id` 关系嵌套
- `recommend_num = 0` 表示该机构没有下级机构

### 树形数据结构示例

```json
[
  {
    "id": 1,
    "parent_id": 0,
    "recommend_num": 555,
    "department_name": "某某大学",
    "department_type": 0,
    "tree_level": 1,
    "items": [
      {
        "id": 5,
        "parent_id": 1,
        "recommend_num": 44,
        "department_name": "学术委员会办公室",
        "department_type": 1,
        "tree_level": 3,
        "items": [
          {
            "id": 6,
            "parent_id": 5,
            "recommend_num": 0,
            "department_name": "学术评审组",
            "department_type": 1,
            "tree_level": 4,
            "items": []
          }
        ]
      },
      {
        "id": 10,
        "parent_id": 1,
        "recommend_num": 120,
        "department_name": "计算机科学与技术学院",
        "department_type": 2,
        "tree_level": 3,
        "items": [
          {
            "id": 11,
            "parent_id": 10,
            "recommend_num": 30,
            "department_name": "软件工程系",
            "department_type": 3,
            "tree_level": 4,
            "items": []
          }
        ]
      }
    ]
  }
]
```

**注意：** 在树形结构中，`tree_level = 3` 的机构（如 id=5 和 id=10）的 `parent_id` 都被修改为学校的 id（1），而不是数据库中原始的 `parent_id` 值。

## 核心查询SQL

### 获取政工人员可见的机构ID列表

```sql
-- 1. 获取政工人员的所有角色
SELECT role_id 
FROM persons_has_roles 
WHERE person_id = {person_id};

-- 2. 获取这些角色的部门权限（department_ids 是 JSON 数组字段）
SELECT department_ids 
FROM persons_roles 
WHERE customer_id = {customer_id} 
  AND id IN (
    SELECT role_id 
    FROM persons_has_roles 
    WHERE person_id = {person_id}
  );

-- 3. 获取直接分配给人员的部门权限
SELECT department_id 
FROM persons_has_department 
WHERE person_id = {person_id};

-- 4. 合并两个来源的部门ID并去重，得到最终的可见部门ID列表
-- 例如：角色权限 [55, 64, 56, 61, 20, 46] + 直接权限 [10, 20, 30] = [55, 64, 56, 61, 20, 46, 10, 30]
```

### 扩展政工可见机构（包含子机构）

```sql
-- 根据政工的部门权限ID，查询这些部门及其所有子部门
SELECT DISTINCT d.id
FROM departments d
WHERE d.customer_id = {customer_id}
  AND d.deleted_at = 0
  AND EXISTS (
    SELECT 1 
    FROM departments p
    WHERE p.id IN ({authorized_dept_ids})  -- 政工有权限的部门ID列表
      AND d.tree_left >= p.tree_left
      AND d.tree_right <= p.tree_right
  );
```

### 获取机构树的查询逻辑

#### 1. 查询学校机构（第一级）

```sql
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_type = 0
  AND deleted_at = 0;
```

#### 2. 查询行政机构（第二级及以下）

```sql
-- 学校管理员：查询所有行政机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_level > 2
  AND department_type = 1
  AND deleted_at = 0
ORDER BY tree_left;

-- 政工人员：仅查询有权限的行政机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_level > 2
  AND department_type = 1
  AND deleted_at = 0
  AND id IN ({visible_dept_ids})  -- 政工可见的机构ID列表（含子机构）
ORDER BY tree_left;
```

#### 3. 查询组织机构（第二级及以下）

```sql
-- 学校管理员：查询所有组织机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_level > 2
  AND department_type != 1
  AND deleted_at = 0
ORDER BY tree_left;

-- 政工人员：仅查询有权限的组织机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_level > 2
  AND department_type != 1
  AND deleted_at = 0
  AND id IN ({visible_dept_ids})  -- 政工可见的机构ID列表（含子机构）
ORDER BY tree_left;
```

### 机构搜索查询

#### 按名称模糊查询

```sql
-- 学校管理员：查询所有机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_name LIKE '%关键词%'
  AND deleted_at = 0
ORDER BY tree_level, tree_left;

-- 政工人员：仅查询有权限的机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_name LIKE '%关键词%'
  AND deleted_at = 0
  AND id IN ({visible_dept_ids})  -- 政工可见的机构ID列表（含子机构）
ORDER BY tree_level, tree_left;
```

#### 按类型查询

```sql
-- 学校管理员：查询所有指定类型的机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_type = 2  -- 例如：查询所有学院
  AND deleted_at = 0
ORDER BY tree_level, tree_left;

-- 政工人员：仅查询有权限的指定类型机构
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_type = 2
  AND deleted_at = 0
  AND id IN ({visible_dept_ids})  -- 政工可见的机构ID列表（含子机构）
ORDER BY tree_level, tree_left;
```

#### 组合查询（名称 + 类型）

```sql
-- 学校管理员
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_name LIKE '%关键词%'
  AND department_type = 2
  AND deleted_at = 0
ORDER BY tree_level, tree_left;

-- 政工人员
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_name LIKE '%关键词%'
  AND department_type = 2
  AND deleted_at = 0
  AND id IN ({visible_dept_ids})  -- 政工可见的机构ID列表（含子机构）
ORDER BY tree_level, tree_left;
```



## 业务逻辑实现

### 获取政工人员可见的机构ID列表

```go
/**
 * 获取政工人员可见的机构ID列表（包含子机构）
 */
func GetStaffVisibleDepartmentIds(customerId, personId int) ([]int, error) {
    // 1. 获取政工人员的所有角色
    var roleIds []int
    err := db.Table("persons_has_roles").
        Where("person_id = ?", personId).
        Pluck("role_id", &roleIds).Error
    if err != nil {
        return nil, err
    }
    
    // 2. 合并两个来源的部门权限
    authorizedDeptIds := make(map[int]bool)
    
    // 2.1 获取角色关联的部门权限（department_ids 是 JSON 数组字段）
    if len(roleIds) > 0 {
        var roles []struct {
            DepartmentIds string `gorm:"column:department_ids"`
        }
        err = db.Table("persons_roles").
            Select("department_ids").
            Where("customer_id = ?", customerId).
            Where("id IN ?", roleIds).
            Find(&roles).Error
        if err != nil {
            return nil, err
        }
        
        // 解析角色的部门权限
        for _, role := range roles {
            if role.DepartmentIds != "" {
                var deptIds []int
                json.Unmarshal([]byte(role.DepartmentIds), &deptIds)
                for _, id := range deptIds {
                    authorizedDeptIds[id] = true
                }
            }
        }
    }
    
    // 2.2 获取直接分配给人员的部门权限
    var directDeptIds []int
    err = db.Table("persons_has_department").
        Where("person_id = ?", personId).
        Pluck("department_id", &directDeptIds).Error
    if err != nil {
        return nil, err
    }
    
    // 合并直接分配的部门权限
    for _, id := range directDeptIds {
        authorizedDeptIds[id] = true
    }
    
    if len(authorizedDeptIds) == 0 {
        return []int{}, nil
    }
    
    // 3. 扩展为包含所有子机构的ID列表
    authorizedIds := make([]int, 0, len(authorizedDeptIds))
    for id := range authorizedDeptIds {
        authorizedIds = append(authorizedIds, id)
    }
    
    // 4. 查询这些部门及其所有子部门
    var visibleDeptIds []int
    err = db.Raw(`
        SELECT DISTINCT d.id
        FROM departments d
        WHERE d.customer_id = ?
          AND d.deleted_at = 0
          AND EXISTS (
            SELECT 1 
            FROM departments p
            WHERE p.id IN ?
              AND d.tree_left >= p.tree_left
              AND d.tree_right <= p.tree_right
          )
    `, customerId, authorizedIds).Pluck("id", &visibleDeptIds).Error
    
    return visibleDeptIds, err
}
```

### 构建机构树的伪代码

```go
/**
 * 获取机构树（两级结构）
 * @param customerId 学校ID
 * @param visibleDeptIds 可见的机构ID列表（政工人员传入，学校管理员传 nil）
 */
func GetDepartmentTree(customerId int, visibleDeptIds []int) ([]DepartmentNode, error) {
    // 1. 查询学校（第一级）
    school := querySchool(customerId)
    
    // 2. 查询行政机构（tree_level > 2, department_type = 1）
    adminDepts := queryAdminDepartments(customerId, visibleDeptIds)
    
    // 3. 查询组织机构（tree_level > 2, department_type != 1）
    orgDepts := queryOrgDepartments(customerId, visibleDeptIds)
    
    // 4. 合并所有机构
    allDepts := append(adminDepts, orgDepts...)
    
    // 5. 构建树形结构
    // 5.1 将 tree_level = 3 的机构作为学校的直接子节点
    // 5.2 将 tree_level > 3 的机构按 parent_id 嵌套到对应父节点
    tree := buildTreeStructure(school, allDepts)
    
    return tree, nil
}

/**
 * 查询行政机构（支持权限过滤）
 */
func queryAdminDepartments(customerId int, visibleDeptIds []int) ([]Department, error) {
    query := db.Table("departments").
        Where("customer_id = ?", customerId).
        Where("tree_level > ?", 2).
        Where("department_type = ?", 1).
        Where("deleted_at = 0")
    
    // 政工人员：添加权限过滤
    if visibleDeptIds != nil && len(visibleDeptIds) > 0 {
        query = query.Where("id IN ?", visibleDeptIds)
    }
    
    var depts []Department
    err := query.Order("tree_left").Find(&depts).Error
    return depts, err
}

/**
 * 查询组织机构（支持权限过滤）
 */
func queryOrgDepartments(customerId int, visibleDeptIds []int) ([]Department, error) {
    query := db.Table("departments").
        Where("customer_id = ?", customerId).
        Where("tree_level > ?", 2).
        Where("department_type != ?", 1).
        Where("deleted_at = 0")
    
    // 政工人员：添加权限过滤
    if visibleDeptIds != nil && len(visibleDeptIds) > 0 {
        query = query.Where("id IN ?", visibleDeptIds)
    }
    
    var depts []Department
    err := query.Order("tree_left").Find(&depts).Error
    return depts, err
}

/**
 * 构建树形结构
 */
func buildTreeStructure(school Department, depts []Department) []DepartmentNode {
    // 创建机构映射表
    deptMap := make(map[int]*DepartmentNode)
    
    // 初始化学校节点
    schoolNode := &DepartmentNode{
        ID:             school.ID,
        ParentID:       school.ParentID,
        RecommendNum:   school.RecommendNum,
        DepartmentName: school.DepartmentName,
        DepartmentType: school.DepartmentType,
        TreeLevel:      school.TreeLevel,
        Items:          []DepartmentNode{},
    }
    deptMap[school.ID] = schoolNode
    
    // 将所有机构加入映射表
    for _, dept := range depts {
        node := &DepartmentNode{
            ID:             dept.ID,
            ParentID:       dept.ParentID,
            RecommendNum:   dept.RecommendNum,
            DepartmentName: dept.DepartmentName,
            DepartmentType: dept.DepartmentType,
            TreeLevel:      dept.TreeLevel,
            Items:          []DepartmentNode{},
        }
        
        // 重要：tree_level = 3 的机构，将 parent_id 改为学校的 id
        if dept.TreeLevel == 3 {
            node.ParentID = school.ID
        }
        
        deptMap[dept.ID] = node
    }
    
    // 构建树形关系
    for _, dept := range depts {
        node := deptMap[dept.ID]
        
        if dept.TreeLevel == 3 {
            // tree_level = 3 的机构直接挂在学校下（parent_id 已改为学校 id）
            schoolNode.Items = append(schoolNode.Items, *node)
        } else if dept.TreeLevel > 3 {
            // tree_level > 3 的机构按原始 parent_id 挂在对应父节点下
            if parentNode, exists := deptMap[dept.ParentID]; exists {
                parentNode.Items = append(parentNode.Items, *node)
            }
        }
    }
    
    return []DepartmentNode{*schoolNode}
}
```

## 缓存策略

### 机构树缓存

**缓存键：** `dept:tree:{customer_id}`

**缓存时长：** 1小时

```go
/**
 * 获取机构树（带缓存）
 */
func GetDepartmentTreeWithCache(customerId int) ([]DepartmentNode, error) {
    cacheKey := fmt.Sprintf("dept:tree:%d", customerId)
    
    // 尝试从缓存获取
    cached, err := redis.Get(cacheKey)
    if err == nil && cached != "" {
        var tree []DepartmentNode
        json.Unmarshal([]byte(cached), &tree)
        return tree, nil
    }
    
    // 查询数据库构建树
    tree, err := GetDepartmentTree(customerId, nil)
    if err != nil {
        return nil, err
    }
    
    // 缓存1小时
    treeJSON, _ := json.Marshal(tree)
    redis.Setex(cacheKey, 3600, string(treeJSON))
    
    return tree, nil
}
```

### 政工可见部门缓存

**说明：** 政工人员的可见部门列表已在登录时缓存到 `staff:managed_scope:{customer_id}:{person_id}`，包含 `managed_department_ids` 字段

**使用方式：**
```go
/**
 * 获取政工可见的部门ID列表（从缓存读取）
 */
func GetStaffVisibleDepartmentIdsFromCache(customerId, personId int) ([]int, error) {
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    
    // 从缓存获取
    cached, err := redis.Get(cacheKey)
    if err != nil || cached == "" {
        // 缓存不存在，需要重新登录或重新计算
        return nil, fmt.Errorf("staff managed scope cache not found")
    }
    
    var scope struct {
        ManagedDepartmentIds []int `json:"managed_department_ids"`
        ManagedPersonIds     []int `json:"managed_person_ids"`
    }
    json.Unmarshal([]byte(cached), &scope)
    
    return scope.ManagedDepartmentIds, nil
}
```

### 缓存清除策略

**清除时机：** 当部门结构或人员权限发生变更时

```go
/**
 * 场景1：机构信息变更时，清除机构树缓存
 * 触发时机：departments 表 INSERT/UPDATE/DELETE
 */
func OnDepartmentChanged(customerId int) {
    // 清除机构树缓存
    cacheKey := fmt.Sprintf("dept:tree:%d", customerId)
    redis.Del(cacheKey)
}

/**
 * 场景2：机构结构变更时，清除机构树缓存 + 清除相关政工的权限缓存
 * 触发时机：departments 表的树结构字段变更（tree_left, tree_right, parent_id）
 */
func OnDepartmentStructureChanged(customerId int, changedDeptIds []int) {
    // 1. 清除机构树缓存
    cacheKey := fmt.Sprintf("dept:tree:%d", customerId)
    redis.Del(cacheKey)
    
    // 2. 清除受影响的政工权限缓存
    // 方案A：简单粗暴，清除所有政工的缓存
    pattern := fmt.Sprintf("staff:managed_scope:%d:*", customerId)
    redis.DelPattern(pattern)
    
    // 方案B：精确清除（需要查询哪些政工管辖了变更的部门）
    // 查询管辖了这些部门的政工
    var affectedPersonIds []int
    db.Raw(`
        SELECT DISTINCT phr.person_id
        FROM persons_has_roles phr
        JOIN persons_roles pr ON phr.role_id = pr.id
        WHERE pr.customer_id = ?
          AND JSON_OVERLAPS(pr.department_ids, ?)
    `, customerId, changedDeptIds).Pluck("person_id", &affectedPersonIds)
    
    // 也查询直接分配了这些部门权限的政工
    var directPersonIds []int
    db.Table("persons_has_department").
        Where("department_id IN ?", changedDeptIds).
        Pluck("person_id", &directPersonIds)
    
    // 合并并清除缓存
    allPersonIds := append(affectedPersonIds, directPersonIds...)
    for _, personId := range allPersonIds {
        cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
        redis.Del(cacheKey)
    }
}

/**
 * 场景3：人员部门权限变更时，清除该人员的权限缓存
 * 触发时机：persons_has_roles、persons_roles、persons_has_department 表变更
 * 注意：此函数在 PersonService 中已定义，这里仅说明关联关系
 */
func OnPersonPermissionChanged(customerId, personId int) {
    // 清除政工权限缓存（包含 managed_department_ids）
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    redis.Del(cacheKey)
}

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 机构树查询 | < 50ms | 1000+ | Redis 1小时（变更时清除） |
| 机构列表查询 | < 30ms | 500+ | 无缓存 |
| 政工可见部门查询 | < 10ms | 2000+ | Redis 永久（登录时缓存，变更时清除） |

## 机构列表查询实现

### 列表接口实现

```go
/**
 * 查询机构列表（支持权限过滤）
 * @param customerId 学校ID
 * @param keyword 机构名称关键词（可选）
 * @param departmentType 机构类型（可选）
 * @param visibleDeptIds 可见的机构ID列表（政工人员传入，学校管理员传 nil）
 */
func GetDepartmentList(customerId int, keyword string, departmentType *int, visibleDeptIds []int) ([]Department, error) {
    query := db.Table("departments").
        Where("customer_id = ?", customerId).
        Where("deleted_at = 0")
    
    // 名称模糊查询
    if keyword != "" {
        query = query.Where("department_name LIKE ?", "%"+keyword+"%")
    }
    
    // 类型查询
    if departmentType != nil {
        query = query.Where("department_type = ?", *departmentType)
    }
    
    // 政工人员：添加权限过滤
    if visibleDeptIds != nil && len(visibleDeptIds) > 0 {
        query = query.Where("id IN ?", visibleDeptIds)
    }
    
    var departments []Department
    err := query.Order("tree_level, tree_left").Find(&departments).Error
    
    return departments, err
}
```

## 数据模型

```go
type Department struct {
    ID             int       `json:"id" gorm:"primaryKey"`
    CustomerID     int       `json:"customer_id"`
    ParentID       int       `json:"parent_id"`
    DepartmentName string    `json:"department_name"`
    DepartmentType int       `json:"department_type"` // 0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级
    RecommendNum   int       `json:"recommend_num"`   // 下级机构数量
    TreeLeft       int       `json:"tree_left"`
    TreeRight      int       `json:"tree_right"`
    TreeLevel      int       `json:"tree_level"`
    Sort           int       `json:"sort"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
    DeletedAt      int       `json:"deleted_at"`
}

type DepartmentNode struct {
    ID             int              `json:"id"`
    ParentID       int              `json:"parent_id"`
    RecommendNum   int              `json:"recommend_num"`
    DepartmentName string           `json:"department_name"`
    DepartmentType int              `json:"department_type"`
    TreeLevel      int              `json:"tree_level"`
    Items          []DepartmentNode `json:"items"`
}
```
