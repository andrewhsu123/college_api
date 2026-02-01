# API 接口参考文档

## 基础信息

- **Base URL**: `http://localhost:8081`
- **认证方式**: Bearer Token
- **响应格式**: JSON

## 认证接口

### 1. 获取学校管理员信息

```
GET /api/base/info
Authorization: Bearer {admin_token}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "mobile": "13800138000",
    "avatar": "",
    "university_id": 1,
    "university_name": "某某大学",
    "status": 1
  }
}
```

### 2. 获取政工人员信息

```
GET /api/staff/info
Authorization: Bearer {staff_token}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "person_id": 38,
    "person_type": 2,
    "name": "李四",
    "gender": 1,
    "mobile": "13900139000",
    "email": "lisi@example.com",
    "avatar": "",
    "status": 1,
    "staff_no": "S001",
    "university_id": 1,
    "university_name": "某某大学",
    "department_id": 5,
    "department_name": "学术委员会办公室",
    "college_id": null,
    "college_name": null,
    "faculty_id": null,
    "faculty_name": null,
    "managed_department_ids": [13, 20, 21, 46, 47, 48, 51, 52, 55, 61, 62, 63, 64],
    "managed_person_ids": [38, 37]
  }
}
```

## 机构接口

### 3. 获取机构树（学校管理员）

```
GET /api/base/departments/tree?customer_id=1
Authorization: Bearer {admin_token}
```

**响应示例**:
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

### 4. 获取机构树（政工人员）

```
GET /api/staff/departments/tree?customer_id=1
Authorization: Bearer {staff_token}
```

**说明**: 仅返回政工人员有权限查看的机构

### 5. 搜索机构列表（学校管理员）

```
GET /api/base/departments/list?customer_id=1&keyword=学院&department_type=2
Authorization: Bearer {admin_token}
```

**参数**:
- `customer_id` (必填): 学校ID
- `keyword` (可选): 机构名称关键词
- `department_type` (可选): 机构类型（0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级）

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 10,
      "parent_id": 1,
      "recommend_num": 120,
      "department_name": "计算机科学与技术学院",
      "department_type": 2,
      "tree_level": 3
    }
  ]
}
```

### 6. 搜索机构列表（政工人员）

```
GET /api/staff/departments/list?customer_id=1&keyword=学院
Authorization: Bearer {staff_token}
```

**说明**: 仅返回政工人员有权限查看的机构

## 人员接口

### 7. 查询人员列表（学校管理员）

```
GET /api/base/persons/list?university_id=1&person_type=1&page=1&page_size=20
Authorization: Bearer {admin_token}
```

**必填参数**:
- `university_id`: 学校ID
- `person_type`: 人员类型（1=学生 2=政工）

**分页参数**:
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）
- `with_extend`: 是否包含扩展信息（默认false）

**基础字段搜索**:
- `name`: 姓名模糊查询
- `mobile`: 手机号
- `email`: 邮箱
- `gender`: 性别（1=男 2=女）
- `status`: 状态（1=正常 2=禁用）

**学生扩展字段搜索**:
- `student_no`: 学号
- `area_id`: 校区ID
- `grade`: 年级
- `education_level`: 教育层次
- `school_system`: 学制
- `id_card`: 身份证号
- `admission_no`: 录取编号
- `exam_no`: 准考证号
- `enrollment_status`: 学籍状态
- `is_enrolled`: 是否报到
- `college_id`: 学院ID
- `faculty_id`: 系ID
- `profession_id`: 专业ID
- `class_id`: 班级ID

**政工扩展字段搜索**:
- `staff_no`: 工号
- `department_id`: 部门ID
- `college_id`: 学院ID
- `faculty_id`: 系ID

**响应示例（不含扩展）**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 10,
        "customer_id": 1,
        "person_type": 1,
        "name": "张三",
        "gender": 1,
        "mobile": "13800138000",
        "email": "zhangsan@example.com",
        "avatar": "https://example.com/avatar/10.jpg",
        "status": 1
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

**响应示例（含学生扩展）**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 10,
        "customer_id": 1,
        "person_type": 1,
        "name": "张三",
        "gender": 1,
        "mobile": "13800138000",
        "email": "zhangsan@example.com",
        "avatar": "https://example.com/avatar/10.jpg",
        "status": 1,
        "student_extend": {
          "person_id": 10,
          "area_id": 1,
          "student_no": "2023001",
          "grade": "2023",
          "education_level": "本科",
          "school_system": "4年",
          "id_card": "110101199001011234",
          "admission_no": "A2023001",
          "exam_no": "E2023001",
          "enrollment_status": 1,
          "is_enrolled": 1,
          "college_id": 10,
          "faculty_id": 11,
          "profession_id": 12,
          "class_id": 13
        }
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

**响应示例（含政工扩展）**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 38,
        "customer_id": 1,
        "person_type": 2,
        "name": "李四",
        "gender": 1,
        "mobile": "13900139000",
        "email": "lisi@example.com",
        "avatar": "https://example.com/avatar/38.jpg",
        "status": 1,
        "staff_extend": {
          "person_id": 38,
          "staff_no": "S001",
          "department_id": 5,
          "college_id": 10,
          "faculty_id": null
        }
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

### 8. 查询人员列表（政工人员）

```
GET /api/staff/persons/list?university_id=1&person_type=1&page=1&page_size=20
Authorization: Bearer {staff_token}
```

**说明**: 
- 参数与学校管理员接口相同
- 仅返回政工人员有权限查看的人员
- 权限过滤规则：
  1. 人员ID在 managed_person_ids 中
  2. 人员所属部门在 managed_department_ids 中

## 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 无权限 |
| 1004 | 资源不存在 |
| 1005 | 内部错误 |

## 使用示例

### 示例1: 查询2023级学生

```bash
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&grade=2023&with_extend=true"
```

### 示例2: 按姓名搜索政工

```bash
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=2&name=张&with_extend=true"
```

### 示例3: 查询某学院的学生

```bash
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&college_id=10&with_extend=true"
```

### 示例4: 政工查询管辖的学生

```bash
curl -H "Authorization: Bearer {staff_token}" \
  "http://localhost:8081/api/staff/persons/list?university_id=1&person_type=1&page=1&page_size=20&with_extend=true"
```

### 示例5: 组合查询（姓名+年级+学院）

```bash
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&name=张&grade=2023&college_id=10&with_extend=true"
```
