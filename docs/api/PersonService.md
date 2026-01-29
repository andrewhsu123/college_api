# PersonService API 接口设计文档

## 1. 概述

**服务名称**：PersonService（人员服务）  
**版本**：v1.0  
**基础路径**：`/api/v1/persons`  
**负责人**：架构师  
**设计日期**：2026-02-09

### 1.1 服务职责

- 人员基础信息的 CRUD 操作
- 人员搜索（姓名、手机号、学号、工号等）
- 人员角色管理
- 人员机构查询
- 人员批量操作

### 1.2 技术栈

- 开发语言：Go 1.21+
- Web框架：Gin
- ORM：GORM
- 缓存：Redis
- 数据库：MySQL 8.0
- 文档：Swagger/OpenAPI 3.0

---

## 2. 数据模型

### 2.1 Person（人员基础信息）

```go
type Person struct {
    ID         uint      `json:"id" gorm:"primaryKey"`
    CustomerID uint      `json:"customer_id" gorm:"index:idx_customer_type"`
    PersonType int8      `json:"person_type" gorm:"index:idx_customer_type;comment:1=学生 2=政工 3=维修工"`
    Name       string    `json:"name" gorm:"size:100"`
    Gender     int8      `json:"gender" gorm:"comment:1=男 2=女"`
    Mobile     string    `json:"mobile" gorm:"size:30"`
    Email      string    `json:"email" gorm:"size:100"`
    Password   string    `json:"-" gorm:"size:255"`
    Avatar     string    `json:"avatar" gorm:"size:500"`
    Status     int8      `json:"status" gorm:"default:1;comment:1=正常 2=禁用"`
    CreatedAt  int64     `json:"created_at"`
    UpdatedAt  int64     `json:"updated_at"`
    DeletedAt  int64     `json:"deleted_at" gorm:"index"`
}
```

### 2.2 PersonDetail（人员详情，包含关联信息）

```go
type PersonDetail struct {
    Person
    Student      *Student      `json:"student,omitempty"`       // 学生信息
    Staff        *Staff        `json:"staff,omitempty"`         // 政工信息
    Roles        []Role        `json:"roles,omitempty"`         // 角色列表
    Departments  []Department  `json:"departments,omitempty"`   // 所属机构
}
```

---

## 3. API 接口列表

### 3.1 人员基础接口

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| GET | /api/v1/persons/:id | 获取人员详情 | 需要权限 |
| GET | /api/v1/persons | 获取人员列表 | 需要权限 |
| POST | /api/v1/persons | 创建人员 | 管理员 |
| PUT | /api/v1/persons/:id | 更新人员 | 管理员 |
| DELETE | /api/v1/persons/:id | 删除人员（软删除） | 管理员 |

### 3.2 人员搜索接口

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| GET | /api/v1/persons/search | 搜索人员（ES） | 需要权限 |

### 3.3 人员角色管理接口

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| GET | /api/v1/persons/:id/roles | 获取人员的所有角色 | 需要权限 |
| POST | /api/v1/persons/:id/roles | 为人员分配角色 | 管理员 |
| DELETE | /api/v1/persons/:id/roles/:role_id | 移除人员角色 | 管理员 |

### 3.4 人员机构查询接口

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| GET | /api/v1/persons/:id/departments | 获取人员所属机构 | 需要权限 |

---

## 4. 接口详细设计

### 4.1 获取人员详情

**接口地址**：`GET /api/v1/persons/:id`

**请求参数**：

| 参数 | 类型 | 位置 | 必填 | 说明 |
|------|------|------|------|------|
| id | int | path | 是 | 人员ID |
| include | string | query | 否 | 包含关联信息，多个用逗号分隔：student,staff,roles,departments |

**请求示例**：
```bash
GET /api/v1/persons/123?include=student,roles
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "customer_id": 1,
    "person_type": 1,
    "name": "张三",
    "mobile": "13800138000",
    "email": "zhangsan@example.com",
    "avatar": "https://example.com/avatar.jpg",
    "gender": 1,
    "status": 1,
    "created_at": 1640000000,
    "updated_at": 1640000000,
    "student": {
      "id": 1,
      "person_id": 123,
      "student_no": "2021001",
      "class_id": 55,
      "college_id": 10,
      "profession_id": 20,
      "faculty_id": 15,
      "grade": 2021,
      "enrollment_status": 1
    },
    "roles": [
      {
        "id": 1,
        "name": "班长",
        "description": "班级管理员"
      }
    ]
  }
}
```

**缓存策略**：
- 缓存Key：`person:info:{person_id}`
- 缓存时间：1小时
- 更新策略：人员信息变更时删除缓存

---

### 4.2 获取人员列表

**接口地址**：`GET /api/v1/persons`

**请求参数**：

| 参数 | 类型 | 位置 | 必填 | 说明 |
|------|------|------|------|------|
| customer_id | int | query | 是 | 客户ID |
| person_type | int | query | 否 | 人员类型：1=学生 2=政工 3=维修工 |
| status | int | query | 否 | 状态：1=正常 2=禁用 |
| keyword | string | query | 否 | 关键词（姓名/手机号） |
| page | int | query | 否 | 页码，默认1 |
| page_size | int | query | 否 | 每页数量，默认20，最大100 |
| order_by | string | query | 否 | 排序字段，默认id |
| order | string | query | 否 | 排序方式：asc/desc，默认desc |

**请求示例**：
```bash
GET /api/v1/persons?customer_id=1&person_type=1&status=1&page=1&page_size=20
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 123,
        "customer_id": 1,
        "person_type": 1,
        "name": "张三",
        "mobile": "13800138000",
        "email": "zhangsan@example.com",
        "avatar": "https://example.com/avatar.jpg",
        "gender": 1,
        "status": 1,
        "created_at": 1640000000,
        "updated_at": 1640000000
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

**性能优化**：
- 使用索引：`idx_customer_type_status`
- 分页优化：使用游标分页（当page较大时）
- 避免 SELECT *，只查询需要的字段

---

### 4.3 创建人员

**接口地址**：`POST /api/v1/persons`

**请求参数**：

| 参数 | 类型 | 位置 | 必填 | 说明 |
|------|------|------|------|------|
| customer_id | int | body | 是 | 客户ID |
| person_type | int | body | 是 | 人员类型：1=学生 2=政工 3=维修工 |
| name | string | body | 是 | 姓名，最大50字符 |
| mobile | string | body | 是 | 手机号，需验证格式 |
| email | string | body | 否 | 邮箱，需验证格式 |
| password | string | body | 是 | 密码，需加密存储 |
| avatar | string | body | 否 | 头像URL |
| gender | int | body | 否 | 性别：1=男 2=女 |
| status | int | body | 否 | 状态：1=正常 2=禁用，默认1 |

**请求示例**：
```json
{
  "customer_id": 1,
  "person_type": 1,
  "name": "张三",
  "mobile": "13800138000",
  "email": "zhangsan@example.com",
  "password": "encrypted_password_here",
  "gender": 1,
  "status": 1
}
```

**响应示例**：
```json
{
  "code": 0,
  "message": "创建成功",
  "data": {
    "id": 123,
    "customer_id": 1,
    "person_type": 1,
    "name": "张三",
    "mobile": "13800138000",
    "email": "zhangsan@example.com",
    "gender": 1,
    "status": 1,
    "created_at": 1640000000,
    "updated_at": 1640000000
  }
}
```

**业务逻辑**：
1. 验证手机号格式和唯一性
2. 验证邮箱格式和唯一性（如果提供）
3. 创建人员记录
4. 发送MQ消息同步到ES
5. 返回创建结果

**错误码**：
- 40001：参数验证失败
- 40002：手机号已存在
- 40003：邮箱已存在

---

### 4.4 更新人员

**接口地址**：`PUT /api/v1/persons/:id`

**请求参数**：

| 参数 | 类型 | 位置 | 必填 | 说明 |
|------|------|------|------|------|
| id | int | path | 是 | 人员ID |
| name | string | body | 否 | 姓名 |
| mobile | string | body | 否 | 手机号 |
| email | string | body | 否 | 邮箱 |
| password | string | body | 否 | 密码（如需修改） |
| avatar | string | body | 否 | 头像URL |
| gender | int | body | 否 | 性别 |
| status | int | body | 否 | 状态 |

**请求示例**：
```json
{
  "name": "张三",
  "mobile": "13800138001",
  "status": 1
}
```

**响应示例**：
```json
{
  "code": 0,
  "message": "更新成功",
  "data": {
    "id": 123,
    "customer_id": 1,
    "person_type": 1,
    "name": "张三",
    "mobile": "13800138001",
    "status": 1,
    "updated_at": 1640000100
  }
}
```

**业务逻辑**：
1. 验证人员是否存在
2. 验证手机号/邮箱唯一性（如果修改）
3. 更新人员信息
4. 删除缓存（延迟双删）
5. 发送MQ消息同步到ES
6. 返回更新结果

---

### 4.5 删除人员

**接口地址**：`DELETE /api/v1/persons/:id`

**请求参数**：

| 参数 | 类型 | 位置 | 必填 | 说明 |
|------|------|------|------|------|
| id | int | path | 是 | 人员ID |

**请求示例**：
```bash
DELETE /api/v1/persons/123
```

**响应示例**：
```json
{
  "code": 0,
  "message": "删除成功",
  "data": null
}
```

**业务逻辑**：
1. 验证人员是否存在
2. 软删除人员（设置 deleted_at）
3. 删除缓存
4. 发送MQ消息从ES删除
5. 返回删除结果

**注意事项**：
- 软删除，不物理删除数据
- 删除前需检查是否有关联数据（角色、机构等）
- 可选：提供强制删除参数

---

