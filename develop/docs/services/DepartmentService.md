# DepartmentService - 机构服务

## 职责

- 机构树查询（学校 -> 行政机构/组织机构的两级树形结构）
- 机构搜索（名称模糊查询、类型查询）

## API 接口

```
GET    /api/departments/tree                # 获取机构树（两级结构）
  参数：customer_id
  返回：树形结构数据
  说明：第一级为学校，第二级为行政机构和组织机构
  
GET    /api/departments/list                # 搜索机构列表
  参数：customer_id, keyword（机构名称模糊查询，可选）, department_type（机构类型，可选）
  返回：机构列表数组
  说明：支持名称模糊查询和类型筛选
```

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
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_level > 2
  AND department_type = 1
  AND deleted_at = 0
ORDER BY tree_left;
```

#### 3. 查询组织机构（第二级及以下）

```sql
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND tree_level > 2
  AND department_type != 1
  AND deleted_at = 0
ORDER BY tree_left;
```

### 机构搜索查询

#### 按名称模糊查询

```sql
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_name LIKE '%关键词%'
  AND deleted_at = 0
ORDER BY tree_level, tree_left;
```

#### 按类型查询

```sql
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_type = 2  -- 例如：查询所有学院
  AND deleted_at = 0
ORDER BY tree_level, tree_left;
```

#### 组合查询（名称 + 类型）

```sql
SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
FROM departments
WHERE customer_id = 1
  AND department_name LIKE '%关键词%'
  AND department_type = 2
  AND deleted_at = 0
ORDER BY tree_level, tree_left;
```



## 业务逻辑实现

### 构建机构树的伪代码

```go
/**
 * 获取机构树（两级结构）
 */
func GetDepartmentTree(customerId int) ([]DepartmentNode, error) {
    // 1. 查询学校（第一级）
    school := querySchool(customerId)
    
    // 2. 查询行政机构（tree_level > 2, department_type = 1）
    adminDepts := queryAdminDepartments(customerId)
    
    // 3. 查询组织机构（tree_level > 2, department_type != 1）
    orgDepts := queryOrgDepartments(customerId)
    
    // 4. 合并所有机构
    allDepts := append(adminDepts, orgDepts...)
    
    // 5. 构建树形结构
    // 5.1 将 tree_level = 3 的机构作为学校的直接子节点
    // 5.2 将 tree_level > 3 的机构按 parent_id 嵌套到对应父节点
    tree := buildTreeStructure(school, allDepts)
    
    return tree, nil
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
    tree, err := GetDepartmentTree(customerId)
    if err != nil {
        return nil, err
    }
    
    // 缓存1小时
    treeJSON, _ := json.Marshal(tree)
    redis.Setex(cacheKey, 3600, string(treeJSON))
    
    return tree, nil
}
```

### 缓存更新

```go
/**
 * 机构变更时
 */
func OnDepartmentChanged(customerId int) {
    // 删除机构树缓存
    cacheKey := fmt.Sprintf("dept:tree:%d", customerId)
    redis.Del(cacheKey)
}
```

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 机构树查询 | < 50ms | 1000+ | Redis 1小时 |
| 机构列表查询 | < 30ms | 500+ | 无缓存 |

## 机构列表查询实现

### 列表接口实现

```go
/**
 * 查询机构列表
 */
func GetDepartmentList(customerId int, keyword string, departmentType *int) ([]Department, error) {
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
