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

### 2.2 PersonInfo（人员完整信息，根据person_type返回不同字段）

```go
// PersonInfo 人员完整信息（根据person_type返回不同字段）
type PersonInfo struct {
    // Person 基础信息
    PersonID       int    `json:"person_id"`
    PersonType     int    `json:"person_type"`
    UniversityID   int    `json:"university_id"`
    UniversityName string `json:"university_name"` // 学校名称
    Name           string `json:"name"`
    Gender         *int   `json:"gender"`
    Mobile         string `json:"mobile"`
    Email          string `json:"email"`
    Avatar         string `json:"avatar"`
    Status         int    `json:"status"` // 1=正常 2=禁用

    // Staff 扩展信息（政工和维修工 person_type=2,3）
    StaffNo        string  `json:"staff_no,omitempty"`
    DepartmentID   *int    `json:"department_id,omitempty"`
    DepartmentName *string `json:"department_name,omitempty"` // 部门名称
    CollegeID      *int    `json:"college_id,omitempty"`
    CollegeName    *string `json:"college_name,omitempty"` // 学院名称
    FacultyID      *int    `json:"faculty_id,omitempty"`
    FacultyName    *string `json:"faculty_name,omitempty"` // 系名称

    // Student 扩展信息（学生 person_type=1）
    StudentNo        string  `json:"student_no,omitempty"`
    Grade            string  `json:"grade,omitempty"`
    AreaID           *int    `json:"area_id,omitempty"`
    EducationLevel   string  `json:"education_level,omitempty"`
    SchoolSystem     string  `json:"school_system,omitempty"`
    IDCard           string  `json:"id_card,omitempty"`
    AdmissionNo      string  `json:"admission_no,omitempty"`
    ExamNo           string  `json:"exam_no,omitempty"`
    EnrollmentStatus *int    `json:"enrollment_status,omitempty"`
    IsEnrolled       *int    `json:"is_enrolled,omitempty"`
    ProfessionID     *int    `json:"profession_id,omitempty"`
    ProfessionName   *string `json:"profession_name,omitempty"` // 专业名称
    ClassID          *int    `json:"class_id,omitempty"`
    ClassName        *string `json:"class_name,omitempty"` // 班级名称

    // 权限信息
    ManagedRoles []ManagedRole `json:"managed_roles"` // 管辖角色及机构
    ManagedMenu  []int         `json:"managed_menu"`  // 菜单权限ID数组
}
```

### 2.3 ManagedRole（管辖角色信息）

```go
// ManagedRole 管辖角色信息
type ManagedRole struct {
    ID          int                 `json:"id"`          // 角色ID
    ParentID    int                 `json:"parent_id"`   // 上级角色组ID
    ParentName  string              `json:"parent_name"` // 上级角色组名称
    Name        string              `json:"name"`        // 角色名称
    Departments []ManagedDepartment `json:"departments"` // 管辖机构列表
}

// ManagedDepartment 管辖机构信息
type ManagedDepartment struct {
    ID             int    `json:"id"`              // 机构ID
    ParentID       int    `json:"parent_id"`       // 上级机构ID
    DepartmentName string `json:"department_name"` // 机构名称
    DepartmentType int    `json:"department_type"` // 机构类型：0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级
    Status         int    `json:"status"`          // 状态
}
```

### 2.4 PersonsRole（管辖角色表）

```go
// PersonsRole 管辖角色
type PersonsRole struct {
    ID          int    `json:"id" db:"id"`
    CustomerID  int    `json:"customer_id" db:"customer_id"`
    ParentID    int    `json:"parent_id" db:"parent_id"`
    Name        string `json:"name" db:"name"`
    Permissions string `json:"permissions" db:"permissions"` // 菜单权限，逗号分隔的ID
}
```

### 2.5 PersonHasDepartment（人员角色管辖机构关系）

```go
// PersonHasDepartment 人员角色管辖机构关系
type PersonHasDepartment struct {
    CustomerID     int `json:"customer_id" db:"customer_id"`
    PersonsRolesID int `json:"persons_roles_id" db:"persons_roles_id"` // 角色ID
    PersonID       int `json:"person_id" db:"person_id"`
    DepartmentID   int `json:"department_id" db:"department_id"`       // 管辖机构ID
}
```

### 2.6 PersonRole（人员角色关系）

```go
// PersonRole 人员角色关系
type PersonRole struct {
    CustomerID int `json:"customer_id" db:"customer_id"`
    PersonID   int `json:"person_id" db:"person_id"`
    RoleID     int `json:"role_id" db:"role_id"`
}
```

---

## 3. API 接口列表

### 3.1 政工端接口（/api/staff）

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/staff/info | 获取当前登录人员信息 | Bearer Token (学生/政工/维修工) |
| GET | /api/staff/departments/tree | 获取管辖机构树 | Bearer Token |
| GET | /api/staff/departments/list | 获取管辖机构列表 | Bearer Token |
| GET | /api/staff/persons/list | 查询人员列表（带权限过滤） | Bearer Token |

### 3.2 学校后台接口（/api/base）

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/base/info | 获取当前登录学校管理员信息 | Bearer Token (学校管理员) |
| GET | /api/base/departments/tree | 获取学校机构树 | Bearer Token |
| GET | /api/base/departments/list | 获取学校机构列表 | Bearer Token |
| GET | /api/base/persons/list | 查询人员列表（无权限过滤） | Bearer Token |

### 3.3 接口说明

#### 政工端 vs 学校后台

| 功能 | 政工端 (/api/staff) | 学校后台 (/api/base) |
|------|---------------------|----------------------|
| 登录用户 | 学生/政工/维修工 (persons表) | 学校管理员 (admin_users表) |
| 机构查询 | 仅查看管辖权限下的机构 | 查看所有机构 |
| 人员查询 | 仅查看管辖权限下的人员 | 查看所有人员 |
| 权限来源 | persons_has_roles + persons_has_department | 无限制 |

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



---

## 5. 人员登录信息接口

### 5.1 获取当前登录人员信息

**接口地址**：`GET /api/staff/info`

**接口说明**：
- 支持学生(person_type=1)、政工(person_type=2)、维修工(person_type=3)登录
- 根据 person_type 返回不同的扩展字段
- 返回管辖角色和菜单权限信息

**请求参数**：无（通过 Token 获取 person_id）

**请求示例**：
```bash
GET /api/staff/info
Authorization: Bearer {token}
```

**响应示例（政工/维修工）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "person_id": 1,
    "person_type": 2,
    "university_id": 1,
    "university_name": "XX大学",
    "name": "张三",
    "gender": 1,
    "mobile": "13800138000",
    "email": "zhangsan@example.com",
    "avatar": "https://example.com/avatar.jpg",
    "status": 1,
    "staff_no": "S001",
    "department_id": 10,
    "department_name": "学生处",
    "college_id": 20,
    "college_name": "计算机学院",
    "faculty_id": 30,
    "faculty_name": "软件工程系",
    "managed_roles": [
      {
        "id": 1,
        "parent_id": 0,
        "parent_name": "",
        "name": "辅导员",
        "departments": [
          {
            "id": 55,
            "parent_id": 20,
            "department_name": "软件2024-1班",
            "department_type": 5,
            "status": 1
          }
        ]
      }
    ],
    "managed_menu": [1, 2, 3, 4]
  }
}
```

**响应示例（学生）**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "person_id": 2,
    "person_type": 1,
    "university_id": 1,
    "university_name": "XX大学",
    "name": "李四",
    "gender": 1,
    "mobile": "13900139000",
    "email": "lisi@example.com",
    "avatar": "https://example.com/avatar2.jpg",
    "status": 1,
    "student_no": "2024001",
    "grade": "2024",
    "area_id": 1,
    "education_level": "本科",
    "school_system": "4年",
    "id_card": "110101200001011234",
    "admission_no": "A2024001",
    "exam_no": "E2024001",
    "enrollment_status": 1,
    "is_enrolled": 1,
    "college_id": 20,
    "college_name": "计算机学院",
    "faculty_id": 30,
    "faculty_name": "软件工程系",
    "profession_id": 40,
    "profession_name": "软件工程",
    "class_id": 55,
    "class_name": "软件2024-1班",
    "managed_roles": [],
    "managed_menu": []
  }
}
```

**响应字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| person_id | int | 人员ID |
| person_type | int | 人员类型：1=学生 2=政工 3=维修工 |
| university_id | int | 学校ID |
| university_name | string | 学校名称 |
| name | string | 姓名 |
| gender | int | 性别：1=男 2=女 |
| mobile | string | 手机号 |
| email | string | 邮箱 |
| avatar | string | 头像URL |
| status | int | 状态：1=正常 2=禁用 |
| staff_no | string | 工号（政工/维修工） |
| department_id | int | 部门ID（政工/维修工） |
| department_name | string | 部门名称（政工/维修工） |
| college_id | int | 学院ID |
| college_name | string | 学院名称 |
| faculty_id | int | 系ID |
| faculty_name | string | 系名称 |
| student_no | string | 学号（学生） |
| grade | string | 年级（学生） |
| area_id | int | 校区ID（学生） |
| education_level | string | 教育层次（学生） |
| school_system | string | 学制（学生） |
| id_card | string | 身份证号（学生） |
| admission_no | string | 录取编号（学生） |
| exam_no | string | 准考证号（学生） |
| enrollment_status | int | 学籍状态（学生） |
| is_enrolled | int | 是否报到（学生） |
| profession_id | int | 专业ID（学生） |
| profession_name | string | 专业名称（学生） |
| class_id | int | 班级ID（学生） |
| class_name | string | 班级名称（学生） |
| managed_roles | array | 管辖角色列表 |
| managed_menu | array | 菜单权限ID数组 |

**managed_roles 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 角色ID |
| parent_id | int | 上级角色组ID |
| parent_name | string | 上级角色组名称 |
| name | string | 角色名称 |
| departments | array | 管辖机构列表 |

**departments 字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 机构ID |
| parent_id | int | 上级机构ID |
| department_name | string | 机构名称 |
| department_type | int | 机构类型：0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级 |
| status | int | 状态 |

**业务逻辑**：
1. 从 Token 获取 person_id
2. 查询 persons 表获取基础信息和 person_type
3. 根据 person_type 查询扩展信息（students 或 staff 表）
4. 查询 departments 表获取机构名称（学校、部门、学院、系、专业、班级）
5. 查询 persons_has_roles 表获取人员角色
6. 查询 persons_roles 表获取角色详情和菜单权限
7. 查询 persons_has_department 表获取角色对应的管辖机构

**注意事项**：
- 机构名称直接从 departments 表查询，不依赖管辖权限
- 学生登录时 managed_roles 和 managed_menu 可能为空
- 政工/维修工的扩展字段使用 omitempty，学生的扩展字段也使用 