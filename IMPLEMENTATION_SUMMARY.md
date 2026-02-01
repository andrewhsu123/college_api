# 人员服务实现总结

## 已完成功能

### 1. 数据模型扩展 (internal/model/person.go)

新增以下数据模型：

- **PersonListRequest**: 人员列表查询请求参数
  - 支持基础字段搜索（姓名、手机、邮箱、性别、状态）
  - 支持学生扩展字段搜索（学号、年级、学院、专业、班级等）
  - 支持政工扩展字段搜索（工号、部门、学院、系）
  - 支持分页参数（page, page_size）
  - 支持扩展信息开关（with_extend）

- **PersonWithExtend**: 人员信息（含扩展）
  - 基础人员信息
  - 可选的学生扩展信息
  - 可选的政工扩展信息

- **StudentExtend**: 学生扩展信息
- **StaffExtend**: 政工扩展信息
- **PersonListResponse**: 人员列表响应（含分页信息）

### 2. 数据访问层 (internal/repository/person_repository.go)

新增以下方法：

- **GetPersonList**: 查询人员列表（支持权限过滤和多条件搜索）
- **buildPermissionFilter**: 构建政工人员权限过滤条件
- **addBasicFilters**: 添加基础字段过滤条件
- **addExtendFilters**: 添加扩展字段过滤条件
- **addStudentFilters**: 添加学生扩展字段过滤
- **addStaffFilters**: 添加政工扩展字段过滤
- **GetStudentExtendInfo**: 批量查询学生扩展信息
- **GetStaffExtendInfo**: 批量查询政工扩展信息
- **addStudentExtendFilters**: 添加学生扩展信息查询的过滤条件
- **addStaffExtendFilters**: 添加政工扩展信息查询的过滤条件

### 3. 业务逻辑层 (internal/service/person_service.go)

新增方法：

- **GetPersonList**: 查询人员列表
  - 设置默认分页参数
  - 调用仓库层查询人员列表
  - 根据需要批量查询扩展信息
  - 合并扩展信息到人员列表

### 4. HTTP处理器

#### 政工端 (internal/handler/staff/person_handler.go)

- **GetPersonList**: 查询人员列表（带权限过滤）
  - 从上下文获取政工人员权限信息
  - 使用 managed_department_ids 和 managed_person_ids 进行过滤

#### 学校管理员端 (internal/handler/base/person_handler.go)

- **GetPersonList**: 查询人员列表
  - 无权限限制，可查看所有人员

### 5. 路由配置 (main.go)

新增路由：

- `GET /api/staff/persons/list` - 政工人员查询列表
- `GET /api/base/persons/list` - 学校管理员查询列表

## API 接口说明

### 请求参数

```
GET /api/staff/persons/list
GET /api/base/persons/list

必填参数：
- university_id: int - 学校ID
- person_type: int - 人员类型（1=学生 2=政工）

分页参数：
- page: int - 页码（默认1）
- page_size: int - 每页数量（默认20，最大100）
- with_extend: bool - 是否包含扩展信息（默认false）

基础字段搜索：
- name: string - 姓名模糊查询
- mobile: string - 手机号
- email: string - 邮箱
- gender: int - 性别（1=男 2=女）
- status: int - 状态（1=正常 2=禁用）

学生扩展字段搜索：
- student_no: string - 学号
- area_id: int - 校区ID
- grade: string - 年级
- education_level: string - 教育层次
- school_system: string - 学制
- id_card: string - 身份证号
- admission_no: string - 录取编号
- exam_no: string - 准考证号
- enrollment_status: int - 学籍状态
- is_enrolled: int - 是否报到
- college_id: int - 学院ID
- faculty_id: int - 系ID
- profession_id: int - 专业ID
- class_id: int - 班级ID

政工扩展字段搜索：
- staff_no: string - 工号
- department_id: int - 部门ID
- college_id: int - 学院ID
- faculty_id: int - 系ID
```

### 响应格式

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

## 权限控制

### 学校管理员
- 可以查看所有人员
- 无权限限制

### 政工人员
- 只能查看管辖权限下的人员
- 权限来源：
  1. 人员ID在 managed_person_ids 中
  2. 人员所属部门在 managed_department_ids 中（通过学生/政工扩展表关联）

## 性能优化

1. **批量查询扩展信息**：避免 N+1 查询问题
2. **分页限制**：最大每页100条记录
3. **条件过滤**：使用 EXISTS 子查询进行权限过滤
4. **索引建议**：
   - persons 表：(customer_id, person_type, deleted_at)
   - students 表：person_id, college_id, faculty_id, profession_id, class_id
   - staff 表：person_id, department_id, college_id, faculty_id

## 测试建议

### 1. 学校管理员查询测试

```bash
# 查询学生列表（不含扩展）
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&page=1&page_size=20"

# 查询学生列表（含扩展）
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&page=1&page_size=20&with_extend=true"

# 按姓名搜索
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&name=张"

# 按年级搜索学生
curl -H "Authorization: Bearer {admin_token}" \
  "http://localhost:8081/api/base/persons/list?university_id=1&person_type=1&grade=2023&with_extend=true"
```

### 2. 政工人员查询测试

```bash
# 查询管辖的学生列表
curl -H "Authorization: Bearer {staff_token}" \
  "http://localhost:8081/api/staff/persons/list?university_id=1&person_type=1&page=1&page_size=20"

# 查询管辖的政工列表（含扩展）
curl -H "Authorization: Bearer {staff_token}" \
  "http://localhost:8081/api/staff/persons/list?university_id=1&person_type=2&with_extend=true"
```

## 后续优化方向

1. **缓存优化**：对热点查询结果进行缓存
2. **ElasticSearch集成**：支持全文搜索和复杂条件查询
3. **异步导出**：大批量数据导出使用异步任务
4. **读写分离**：查询操作使用从库
