# 机构服务 API 文档

## 概述

机构服务提供组织机构的树形查询和列表查询功能，支持学校管理员和政工人员两种角色的权限控制。

## 接口列表

### 学校管理员接口

#### 1. 获取机构树

```
GET /api/base/departments/tree
```

**认证：** Bearer Token (学校管理员)

**请求参数：** 无（customer_id 从 token 自动获取）

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
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
          "items": []
        }
      ]
    }
  ]
}
```

#### 2. 搜索机构列表

```
GET /api/base/departments/list
```

**认证：** Bearer Token (学校管理员)

**请求参数：**
- `keyword` (可选): 机构名称关键词
- `department_type` (可选): 机构类型 (0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级)

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 5,
      "parent_id": 2,
      "recommend_num": 44,
      "department_name": "学术委员会办公室",
      "department_type": 1,
      "tree_level": 3
    }
  ]
}
```

### 政工人员接口

#### 3. 获取机构树（仅可见有权限的机构）

```
GET /api/staff/departments/tree
```

**认证：** Bearer Token (政工人员)

**请求参数：** 无（customer_id 和 person_id 从 token 自动获取）

**权限说明：** 仅返回政工人员被授权的机构及其子机构

**响应示例：** 同学校管理员接口，但数据已根据权限过滤

#### 4. 搜索机构列表（仅可见有权限的机构）

```
GET /api/staff/departments/list
```

**认证：** Bearer Token (政工人员)

**请求参数：**
- `keyword` (可选): 机构名称关键词
- `department_type` (可选): 机构类型

**权限说明：** 仅返回政工人员被授权的机构及其子机构

**响应示例：** 同学校管理员接口，但数据已根据权限过滤

## 机构类型说明

| 类型值 | 说明 |
|-------|------|
| 0 | 学校 |
| 1 | 行政机构 |
| 2 | 学院 |
| 3 | 系 |
| 4 | 专业 |
| 5 | 班级 |

## 树形结构说明

- **第一级：** 学校（tree_level = 1）
- **第二级：** 行政机构和组织机构（tree_level = 3）
- **第三级及以下：** 按实际层级嵌套（tree_level > 3）

**注意：** 在树形结构中，tree_level = 3 的机构的 parent_id 会被修改为学校的 id，以实现两级展示效果。

## 权限模型

### 学校管理员
- 登录接口：`GET /api/base/info`
- 权限范围：可以查看所有机构
- Token 类型：`Modules\User\Models\User`

### 政工人员
- 登录接口：`GET /api/staff/info`
- 权限范围：只能查看被授权的机构及其子机构
- Token 类型：`Modules\Persons\Models\Person` 或类似
- 人员类型：`person_type = 2`（政工）

### 权限查询流程

1. 从 `persons_has_roles` 表获取政工人员的角色ID列表
2. 从 `persons_roles` 表获取这些角色的 `department_ids`（JSON数组）
3. 合并去重得到授权的部门ID列表
4. 使用嵌套集合（Nested Set）查询扩展为包含所有子部门的ID列表
5. 在查询时添加 `id IN (...)` 条件过滤

## 测试示例

### 学校管理员测试

```bash
# 获取机构树
curl -X GET "http://127.0.0.1:8081/api/base/departments/tree" \
  -H "Authorization: Bearer 111|cxlho8PuMFjSqUQmWiecuhG2cc3YdVUZ0TK9xF4568f3b872"

# 搜索机构列表
curl -X GET "http://127.0.0.1:8081/api/base/departments/list?keyword=学院&department_type=2" \
  -H "Authorization: Bearer 111|cxlho8PuMFjSqUQmWiecuhG2cc3YdVUZ0TK9xF4568f3b872"
```

### 政工人员测试

```bash
# 获取机构树（仅可见有权限的）
curl -X GET "http://127.0.0.1:8081/api/staff/departments/tree" \
  -H "Authorization: Bearer 112|au4XBYmV0w8IjGD0KviuLsv4X1oBHyJ1LPAXdpIM259050c6"

# 搜索机构列表（仅可见有权限的）
curl -X GET "http://127.0.0.1:8081/api/staff/departments/list?keyword=办公室" \
  -H "Authorization: Bearer 112|au4XBYmV0w8IjGD0KviuLsv4X1oBHyJ1LPAXdpIM259050c6"
```

## 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 401 | 未授权（Token无效或已过期） |
| 500 | 服务器内部错误 |

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 机构树查询 | < 50ms | 1000+ | Redis 1小时（待实现） |
| 机构列表查询 | < 30ms | 500+ | 无缓存 |
