# PersonService - 人员服务

## 职责

- 人员列表查询（支持多条件搜索和分页）
- 人员详情查询（支持扩展信息）
- 支持学校管理员和政工人员两种角色的权限过滤
- 根据人员类型动态加载扩展信息（学生/政工）

## API 接口

### 学校管理员接口（/api/base）

```
GET    /api/base/persons/list              # 查询人员列表
  认证：Bearer Token (学校管理员)
  参数：
    - university_id: int (必填) 学校ID
    - person_type: int (必填) 人员类型 1=学生 2=政工
    - page: int (可选，默认1) 页码
    - page_size: int (可选，默认20) 每页数量
    - with_extend: bool (可选，默认false) 是否包含扩展信息
    - name: string (可选) 姓名模糊查询
    - mobile: string (可选) 手机号
    - email: string (可选) 邮箱
    - gender: int (可选) 性别 1=男 2=女
    - status: int (可选) 状态 1=正常 2=禁用
    
    # 政工扩展字段
    - staff_no: string (可选) 工号
    - department_id: int (可选) 所属部门ID
    - college_id: int (可选) 所属学院ID
    - faculty_id: int (可选) 所属系ID
    
    # 学生扩展字段
    - student_no: string (可选) 学号
    - area_id: int (可选) 校区ID
    - grade: string (可选) 年级
    - education_level: string (可选) 教育层次
    - school_system: string (可选) 学制
    - id_card: string (可选) 身份证号
    - admission_no: string (可选) 录取编号
    - exam_no: string (可选) 准考证号
    - enrollment_status: int (可选) 学籍状态
    - is_enrolled: int (可选) 是否报到
    - profession_id: int (可选) 专业ID
    - class_id: int (可选) 班级ID
  
  返回：分页人员列表
  权限：查看所有人员
```

### 政工人员接口（/api/staff）

```
GET    /api/staff/persons/list             # 查询人员列表
  认证：Bearer Token (政工人员)
  参数：同学校管理员接口
  返回：分页人员列表
  权限：仅查看管辖权限下的人员
    - 人员所属部门在管辖部门下（通过部门树判断）
    - 或人员ID在 managed_person_ids 数组中
```

## 用户角色说明

### 学校管理员
- 登录接口：`GET /api/base/info`
- 认证方式：`Authorization: Bearer {token}`
- 权限范围：可以查看所有人员
- 用户标识：通过 `admin_users` 表的 `user_id` 获取 `customer_id`

```sql
-- 获取学校管理员的 customer_id
SELECT customer_id FROM admin_users WHERE id = {user_id};
```

### 政工人员
- 登录接口：`GET /api/staff/info`
- 认证方式：`Authorization: Bearer {token}`
- 权限范围：只能查看管辖权限下的人员
- 用户标识：通过 `persons` 表的 `person_id` 获取 `customer_id` 和权限

```sql
-- 获取政工人员的 customer_id
SELECT customer_id FROM persons WHERE id = {person_id};

-- 获取政工人员管辖的部门ID列表（两个来源）
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

-- 获取政工人员管辖的人员ID列表
-- 从角色的 person_ids 字段（JSON数组）中获取
SELECT person_ids 
FROM persons_roles 
WHERE customer_id = {customer_id} 
  AND id IN (
    SELECT role_id 
    FROM persons_has_roles 
    WHERE person_id = {person_id}
  );
```

**权限示例：**
```json
// 角色1的权限
{
  "department_ids": [55, 64, 56, 61, 20, 46],
  "person_ids": [38, 37]
}

// 角色2的权限
{
  "department_ids": [55],
  "person_ids": []
}

// 直接分配的部门权限
[10, 20, 30]

// 合并后的最终权限
{
  "managed_department_ids": [55, 64, 56, 61, 20, 46, 10, 30],
  "managed_person_ids": [38, 37]
}
```

## 数据库表

### persons（人员基础表）

查询字段：
- `id` - 人员ID
- `customer_id` - 客户ID(学校ID)
- `person_type` - 人员类型:1=学生 2=政工
- `name` - 姓名
- `gender` - 性别:1=男,2=女
- `mobile` - 手机号
- `email` - 电子邮箱
- `avatar` - 头像
- `status` - 状态:1=正常,2=禁用

### students（学生扩展表）

扩展字段（当 person_type=1 且 with_extend=true 时查询）：
- `area_id` - 校区ID
- `student_no` - 学号
- `grade` - 年级
- `education_level` - 教育层次
- `school_system` - 学制
- `id_card` - 身份证号
- `admission_no` - 录取编号
- `exam_no` - 准考证号
- `enrollment_status` - 学籍状态
- `is_enrolled` - 是否报到
- `college_id` - 学院ID
- `faculty_id` - 系ID
- `profession_id` - 专业ID
- `class_id` - 班级ID

### staff（政工扩展表）

扩展字段（当 person_type=2 且 with_extend=true 时查询）：
- `staff_no` - 工号
- `department_id` - 所属部门ID
- `college_id` - 所属学院ID
- `faculty_id` - 所属系ID

## 核心查询SQL

### 1. 获取政工人员的管辖权限

```sql
-- Step 1: 获取政工人员的所有角色
SELECT role_id 
FROM persons_has_roles 
WHERE person_id = {person_id};

-- Step 2: 获取角色的部门权限和人员权限
SELECT department_ids, person_ids
FROM persons_roles 
WHERE customer_id = {customer_id} 
  AND id IN ({role_ids});

-- Step 3: 获取直接分配的部门权限
SELECT department_id 
FROM persons_has_department 
WHERE person_id = {person_id};

-- Step 4: 扩展部门权限（包含所有子部门）
SELECT DISTINCT d.id
FROM departments d
WHERE d.customer_id = {customer_id}
  AND d.deleted_at = 0
  AND EXISTS (
    SELECT 1 
    FROM departments p
    WHERE p.id IN ({managed_department_ids})
      AND d.tree_left >= p.tree_left
      AND d.tree_right <= p.tree_right
  );
```

### 2. 查询人员列表（学校管理员）

```sql
-- 基础查询（persons表）
SELECT 
  id, customer_id, person_type, name, gender, 
  mobile, email, avatar, status
FROM persons
WHERE customer_id = {university_id}
  AND person_type = {person_type}
  AND deleted_at = 0
  [AND name LIKE '%{keyword}%']           -- 可选：姓名模糊查询
  [AND mobile = '{mobile}']               -- 可选：手机号
  [AND email = '{email}']                 -- 可选：邮箱
  [AND gender = {gender}]                 -- 可选：性别
  [AND status = {status}]                 -- 可选：状态
ORDER BY id DESC
LIMIT {offset}, {page_size};

-- 获取总数
SELECT COUNT(*) 
FROM persons
WHERE customer_id = {university_id}
  AND person_type = {person_type}
  AND deleted_at = 0
  [AND name LIKE '%{keyword}%']
  [AND mobile = '{mobile}']
  [AND email = '{email}']
  [AND gender = {gender}']
  [AND status = {status}];
```

### 3. 查询人员列表（政工人员）

```sql
-- 基础查询（persons表 + 权限过滤）
SELECT 
  p.id, p.customer_id, p.person_type, p.name, p.gender, 
  p.mobile, p.email, p.avatar, p.status
FROM persons p
WHERE p.customer_id = {university_id}
  AND p.person_type = {person_type}
  AND p.deleted_at = 0
  AND (
    -- 条件1：人员ID在管辖人员列表中
    p.id IN ({managed_person_ids})
    OR
    -- 条件2：人员所属部门在管辖部门下
    EXISTS (
      SELECT 1 FROM students s
      WHERE s.person_id = p.id
        AND (
          s.college_id IN ({visible_dept_ids})
          OR s.faculty_id IN ({visible_dept_ids})
          OR s.profession_id IN ({visible_dept_ids})
          OR s.class_id IN ({visible_dept_ids})
        )
    )
    OR
    EXISTS (
      SELECT 1 FROM staff st
      WHERE st.person_id = p.id
        AND (
          st.department_id IN ({visible_dept_ids})
          OR st.college_id IN ({visible_dept_ids})
          OR st.faculty_id IN ({visible_dept_ids})
        )
    )
  )
  [AND p.name LIKE '%{keyword}%']
  [AND p.mobile = '{mobile}']
  [AND p.email = '{email}']
  [AND p.gender = {gender}]
  [AND p.status = {status}]
ORDER BY p.id DESC
LIMIT {offset}, {page_size};
```

### 4. 查询扩展信息（学生）

```sql
-- 批量查询学生扩展信息
SELECT 
  person_id, area_id, student_no, grade, education_level,
  school_system, id_card, admission_no, exam_no,
  enrollment_status, is_enrolled, college_id, faculty_id,
  profession_id, class_id
FROM students
WHERE person_id IN ({person_ids})
  [AND student_no LIKE '%{student_no}%']      -- 可选：学号
  [AND area_id = {area_id}]                   -- 可选：校区
  [AND grade = '{grade}']                     -- 可选：年级
  [AND education_level = '{education_level}'] -- 可选：教育层次
  [AND school_system = '{school_system}']     -- 可选：学制
  [AND id_card = '{id_card}']                 -- 可选：身份证号
  [AND admission_no = '{admission_no}']       -- 可选：录取编号
  [AND exam_no = '{exam_no}']                 -- 可选：准考证号
  [AND enrollment_status = {enrollment_status}] -- 可选：学籍状态
  [AND is_enrolled = {is_enrolled}]           -- 可选：是否报到
  [AND college_id = {college_id}]             -- 可选：学院
  [AND faculty_id = {faculty_id}]             -- 可选：系
  [AND profession_id = {profession_id}]       -- 可选：专业
  [AND class_id = {class_id}];                -- 可选：班级
```

### 5. 查询扩展信息（政工）

```sql
-- 批量查询政工扩展信息
SELECT 
  person_id, staff_no, department_id, college_id, faculty_id
FROM staff
WHERE person_id IN ({person_ids})
  [AND staff_no LIKE '%{staff_no}%']          -- 可选：工号
  [AND department_id = {department_id}]       -- 可选：部门
  [AND college_id = {college_id}]             -- 可选：学院
  [AND faculty_id = {faculty_id}];            -- 可选：系
```

## 业务逻辑实现

### 获取政工人员管辖权限

```go
/**
 * 获取政工人员的管辖权限
 * @return managed_department_ids 管辖的部门ID列表（含子部门）
 * @return managed_person_ids 管辖的人员ID列表
 */
func GetStaffManagedScope(customerId, personId int) ([]int, []int, error) {
    // 1. 获取政工人员的所有角色
    var roleIds []int
    err := db.Table("persons_has_roles").
        Where("person_id = ?", personId).
        Pluck("role_id", &roleIds).Error
    if err != nil {
        return nil, nil, err
    }
    
    // 2. 合并部门权限和人员权限
    managedDeptIds := make(map[int]bool)
    managedPersonIds := make(map[int]bool)
    
    // 2.1 获取角色的权限
    if len(roleIds) > 0 {
        var roles []struct {
            DepartmentIds string `gorm:"column:department_ids"`
            PersonIds     string `gorm:"column:person_ids"`
        }
        err = db.Table("persons_roles").
            Select("department_ids, person_ids").
            Where("customer_id = ?", customerId).
            Where("id IN ?", roleIds).
            Find(&roles).Error
        if err != nil {
            return nil, nil, err
        }
        
        // 解析角色权限
        for _, role := range roles {
            // 解析部门权限
            if role.DepartmentIds != "" {
                var deptIds []int
                json.Unmarshal([]byte(role.DepartmentIds), &deptIds)
                for _, id := range deptIds {
                    managedDeptIds[id] = true
                }
            }
            
            // 解析人员权限
            if role.PersonIds != "" {
                var personIds []int
                json.Unmarshal([]byte(role.PersonIds), &personIds)
                for _, id := range personIds {
                    managedPersonIds[id] = true
                }
            }
        }
    }
    
    // 2.2 获取直接分配的部门权限
    var directDeptIds []int
    err = db.Table("persons_has_department").
        Where("person_id = ?", personId).
        Pluck("department_id", &directDeptIds).Error
    if err != nil {
        return nil, nil, err
    }
    
    for _, id := range directDeptIds {
        managedDeptIds[id] = true
    }
    
    // 3. 扩展部门权限（包含所有子部门）
    if len(managedDeptIds) == 0 {
        // 转换为数组
        personIdsArray := make([]int, 0, len(managedPersonIds))
        for id := range managedPersonIds {
            personIdsArray = append(personIdsArray, id)
        }
        return []int{}, personIdsArray, nil
    }
    
    deptIdsArray := make([]int, 0, len(managedDeptIds))
    for id := range managedDeptIds {
        deptIdsArray = append(deptIdsArray, id)
    }
    
    // 查询这些部门及其所有子部门
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
    `, customerId, deptIdsArray).Pluck("id", &visibleDeptIds).Error
    if err != nil {
        return nil, nil, err
    }
    
    // 转换人员ID为数组
    personIdsArray := make([]int, 0, len(managedPersonIds))
    for id := range managedPersonIds {
        personIdsArray = append(personIdsArray, id)
    }
    
    return visibleDeptIds, personIdsArray, nil
}
```

### 查询人员列表（带权限过滤）

```go
/**
 * 查询人员列表
 * @param req 查询请求参数
 * @param isStaff 是否政工人员
 * @param visibleDeptIds 可见部门ID列表（政工人员）
 * @param managedPersonIds 管辖人员ID列表（政工人员）
 */
func GetPersonList(req PersonListRequest, isStaff bool, visibleDeptIds, managedPersonIds []int) (*PersonListResponse, error) {
    // 1. 构建基础查询
    query := db.Table("persons p").
        Select("p.id, p.customer_id, p.person_type, p.name, p.gender, p.mobile, p.email, p.avatar, p.status").
        Where("p.customer_id = ?", req.UniversityID).
        Where("p.person_type = ?", req.PersonType).
        Where("p.deleted_at = 0")
    
    // 2. 政工人员：添加权限过滤
    if isStaff {
        query = query.Where(func(q *gorm.DB) *gorm.DB {
            subQuery := q
            
            // 条件1：人员ID在管辖列表中
            if len(managedPersonIds) > 0 {
                subQuery = subQuery.Or("p.id IN ?", managedPersonIds)
            }
            
            // 条件2：人员所属部门在管辖部门下
            if len(visibleDeptIds) > 0 {
                if req.PersonType == 1 { // 学生
                    subQuery = subQuery.Or(`EXISTS (
                        SELECT 1 FROM students s
                        WHERE s.person_id = p.id
                          AND (
                            s.college_id IN ?
                            OR s.faculty_id IN ?
                            OR s.profession_id IN ?
                            OR s.class_id IN ?
                          )
                    )`, visibleDeptIds, visibleDeptIds, visibleDeptIds, visibleDeptIds)
                } else if req.PersonType == 2 { // 政工
                    subQuery = subQuery.Or(`EXISTS (
                        SELECT 1 FROM staff st
                        WHERE st.person_id = p.id
                          AND (
                            st.department_id IN ?
                            OR st.college_id IN ?
                            OR st.faculty_id IN ?
                          )
                    )`, visibleDeptIds, visibleDeptIds, visibleDeptIds)
                }
            }
            
            return subQuery
        })
    }
    
    // 3. 添加搜索条件（persons表字段）
    if req.Name != "" {
        query = query.Where("p.name LIKE ?", "%"+req.Name+"%")
    }
    if req.Mobile != "" {
        query = query.Where("p.mobile = ?", req.Mobile)
    }
    if req.Email != "" {
        query = query.Where("p.email = ?", req.Email)
    }
    if req.Gender != nil {
        query = query.Where("p.gender = ?", *req.Gender)
    }
    if req.Status != nil {
        query = query.Where("p.status = ?", *req.Status)
    }
    
    // 4. 添加扩展表的搜索条件
    if req.PersonType == 1 { // 学生
        query = addStudentFilters(query, req)
    } else if req.PersonType == 2 { // 政工
        query = addStaffFilters(query, req)
    }
    
    // 5. 获取总数
    var total int64
    countQuery := query
    err := countQuery.Count(&total).Error
    if err != nil {
        return nil, err
    }
    
    // 6. 分页查询
    offset := (req.Page - 1) * req.PageSize
    var persons []Person
    err = query.Order("p.id DESC").
        Limit(req.PageSize).
        Offset(offset).
        Find(&persons).Error
    if err != nil {
        return nil, err
    }
    
    // 7. 如果需要扩展信息，批量查询
    if req.WithExtend && len(persons) > 0 {
        personIds := make([]int, len(persons))
        for i, p := range persons {
            personIds[i] = p.ID
        }
        
        if req.PersonType == 1 {
            // 查询学生扩展信息
            extendMap, err := getStudentExtendInfo(personIds, req)
            if err != nil {
                return nil, err
            }
            
            // 合并扩展信息
            for i := range persons {
                if extend, ok := extendMap[persons[i].ID]; ok {
                    persons[i].StudentExtend = extend
                }
            }
        } else if req.PersonType == 2 {
            // 查询政工扩展信息
            extendMap, err := getStaffExtendInfo(personIds, req)
            if err != nil {
                return nil, err
            }
            
            // 合并扩展信息
            for i := range persons {
                if extend, ok := extendMap[persons[i].ID]; ok {
                    persons[i].StaffExtend = extend
                }
            }
        }
    }
    
    return &PersonListResponse{
        Items:    persons,
        Total:    total,
        Page:     req.Page,
        PageSize: req.PageSize,
    }, nil
}

/**
 * 添加学生扩展表的过滤条件
 */
func addStudentFilters(query *gorm.DB, req PersonListRequest) *gorm.DB {
    hasStudentFilter := false
    studentQuery := ""
    studentParams := []interface{}{}
    
    if req.StudentNo != "" {
        hasStudentFilter = true
        studentQuery += " AND s.student_no LIKE ?"
        studentParams = append(studentParams, "%"+req.StudentNo+"%")
    }
    if req.AreaID != nil {
        hasStudentFilter = true
        studentQuery += " AND s.area_id = ?"
        studentParams = append(studentParams, *req.AreaID)
    }
    if req.Grade != "" {
        hasStudentFilter = true
        studentQuery += " AND s.grade = ?"
        studentParams = append(studentParams, req.Grade)
    }
    if req.EducationLevel != "" {
        hasStudentFilter = true
        studentQuery += " AND s.education_level = ?"
        studentParams = append(studentParams, req.EducationLevel)
    }
    if req.SchoolSystem != "" {
        hasStudentFilter = true
        studentQuery += " AND s.school_system = ?"
        studentParams = append(studentParams, req.SchoolSystem)
    }
    if req.IDCard != "" {
        hasStudentFilter = true
        studentQuery += " AND s.id_card = ?"
        studentParams = append(studentParams, req.IDCard)
    }
    if req.AdmissionNo != "" {
        hasStudentFilter = true
        studentQuery += " AND s.admission_no = ?"
        studentParams = append(studentParams, req.AdmissionNo)
    }
    if req.ExamNo != "" {
        hasStudentFilter = true
        studentQuery += " AND s.exam_no = ?"
        studentParams = append(studentParams, req.ExamNo)
    }
    if req.EnrollmentStatus != nil {
        hasStudentFilter = true
        studentQuery += " AND s.enrollment_status = ?"
        studentParams = append(studentParams, *req.EnrollmentStatus)
    }
    if req.IsEnrolled != nil {
        hasStudentFilter = true
        studentQuery += " AND s.is_enrolled = ?"
        studentParams = append(studentParams, *req.IsEnrolled)
    }
    if req.CollegeID != nil {
        hasStudentFilter = true
        studentQuery += " AND s.college_id = ?"
        studentParams = append(studentParams, *req.CollegeID)
    }
    if req.FacultyID != nil {
        hasStudentFilter = true
        studentQuery += " AND s.faculty_id = ?"
        studentParams = append(studentParams, *req.FacultyID)
    }
    if req.ProfessionID != nil {
        hasStudentFilter = true
        studentQuery += " AND s.profession_id = ?"
        studentParams = append(studentParams, *req.ProfessionID)
    }
    if req.ClassID != nil {
        hasStudentFilter = true
        studentQuery += " AND s.class_id = ?"
        studentParams = append(studentParams, *req.ClassID)
    }
    
    if hasStudentFilter {
        query = query.Where("EXISTS (SELECT 1 FROM students s WHERE s.person_id = p.id"+studentQuery+")", studentParams...)
    }
    
    return query
}

/**
 * 添加政工扩展表的过滤条件
 */
func addStaffFilters(query *gorm.DB, req PersonListRequest) *gorm.DB {
    hasStaffFilter := false
    staffQuery := ""
    staffParams := []interface{}{}
    
    if req.StaffNo != "" {
        hasStaffFilter = true
        staffQuery += " AND st.staff_no LIKE ?"
        staffParams = append(staffParams, "%"+req.StaffNo+"%")
    }
    if req.DepartmentID != nil {
        hasStaffFilter = true
        staffQuery += " AND st.department_id = ?"
        staffParams = append(staffParams, *req.DepartmentID)
    }
    if req.CollegeID != nil {
        hasStaffFilter = true
        staffQuery += " AND st.college_id = ?"
        staffParams = append(staffParams, *req.CollegeID)
    }
    if req.FacultyID != nil {
        hasStaffFilter = true
        staffQuery += " AND st.faculty_id = ?"
        staffParams = append(staffParams, *req.FacultyID)
    }
    
    if hasStaffFilter {
        query = query.Where("EXISTS (SELECT 1 FROM staff st WHERE st.person_id = p.id"+staffQuery+")", staffParams...)
    }
    
    return query
}
```

### 批量查询扩展信息

```go
/**
 * 批量查询学生扩展信息
 */
func getStudentExtendInfo(personIds []int, req PersonListRequest) (map[int]*StudentExtend, error) {
    query := db.Table("students").
        Where("person_id IN ?", personIds)
    
    // 添加学生扩展字段的过滤条件
    if req.StudentNo != "" {
        query = query.Where("student_no LIKE ?", "%"+req.StudentNo+"%")
    }
    if req.AreaID != nil {
        query = query.Where("area_id = ?", *req.AreaID)
    }
    if req.Grade != "" {
        query = query.Where("grade = ?", req.Grade)
    }
    if req.EducationLevel != "" {
        query = query.Where("education_level = ?", req.EducationLevel)
    }
    if req.SchoolSystem != "" {
        query = query.Where("school_system = ?", req.SchoolSystem)
    }
    if req.IDCard != "" {
        query = query.Where("id_card = ?", req.IDCard)
    }
    if req.AdmissionNo != "" {
        query = query.Where("admission_no = ?", req.AdmissionNo)
    }
    if req.ExamNo != "" {
        query = query.Where("exam_no = ?", req.ExamNo)
    }
    if req.EnrollmentStatus != nil {
        query = query.Where("enrollment_status = ?", *req.EnrollmentStatus)
    }
    if req.IsEnrolled != nil {
        query = query.Where("is_enrolled = ?", *req.IsEnrolled)
    }
    if req.CollegeID != nil {
        query = query.Where("college_id = ?", *req.CollegeID)
    }
    if req.FacultyID != nil {
        query = query.Where("faculty_id = ?", *req.FacultyID)
    }
    if req.ProfessionID != nil {
        query = query.Where("profession_id = ?", *req.ProfessionID)
    }
    if req.ClassID != nil {
        query = query.Where("class_id = ?", *req.ClassID)
    }
    
    var students []StudentExtend
    err := query.Find(&students).Error
    if err != nil {
        return nil, err
    }
    
    // 转换为 map
    extendMap := make(map[int]*StudentExtend)
    for i := range students {
        extendMap[students[i].PersonID] = &students[i]
    }
    
    return extendMap, nil
}

/**
 * 批量查询政工扩展信息
 */
func getStaffExtendInfo(personIds []int, req PersonListRequest) (map[int]*StaffExtend, error) {
    query := db.Table("staff").
        Where("person_id IN ?", personIds)
    
    // 添加政工扩展字段的过滤条件
    if req.StaffNo != "" {
        query = query.Where("staff_no LIKE ?", "%"+req.StaffNo+"%")
    }
    if req.DepartmentID != nil {
        query = query.Where("department_id = ?", *req.DepartmentID)
    }
    if req.CollegeID != nil {
        query = query.Where("college_id = ?", *req.CollegeID)
    }
    if req.FacultyID != nil {
        query = query.Where("faculty_id = ?", *req.FacultyID)
    }
    
    var staffs []StaffExtend
    err := query.Find(&staffs).Error
    if err != nil {
        return nil, err
    }
    
    // 转换为 map
    extendMap := make(map[int]*StaffExtend)
    for i := range staffs {
        extendMap[staffs[i].PersonID] = &staffs[i]
    }
    
    return extendMap, nil
}
```

## 数据模型

```go
// 人员列表请求参数
type PersonListRequest struct {
    // 必填参数
    UniversityID int `form:"university_id" binding:"required"` // 学校ID
    PersonType   int `form:"person_type" binding:"required"`   // 人员类型 1=学生 2=政工
    
    // 分页参数
    Page     int  `form:"page" binding:"min=1"`                // 页码，默认1
    PageSize int  `form:"page_size" binding:"min=1,max=100"`   // 每页数量，默认20
    WithExtend bool `form:"with_extend"`                        // 是否包含扩展信息
    
    // persons表字段搜索
    Name   string `form:"name"`   // 姓名模糊查询
    Mobile string `form:"mobile"` // 手机号
    Email  string `form:"email"`  // 邮箱
    Gender *int   `form:"gender"` // 性别 1=男 2=女
    Status *int   `form:"status"` // 状态 1=正常 2=禁用
    
    // 政工扩展字段搜索
    StaffNo      string `form:"staff_no"`      // 工号
    DepartmentID *int   `form:"department_id"` // 所属部门ID
    
    // 学生扩展字段搜索
    StudentNo        string `form:"student_no"`        // 学号
    AreaID           *int   `form:"area_id"`           // 校区ID
    Grade            string `form:"grade"`             // 年级
    EducationLevel   string `form:"education_level"`   // 教育层次
    SchoolSystem     string `form:"school_system"`     // 学制
    IDCard           string `form:"id_card"`           // 身份证号
    AdmissionNo      string `form:"admission_no"`      // 录取编号
    ExamNo           string `form:"exam_no"`           // 准考证号
    EnrollmentStatus *int   `form:"enrollment_status"` // 学籍状态
    IsEnrolled       *int   `form:"is_enrolled"`       // 是否报到
    
    // 共用的组织字段
    CollegeID    *int `form:"college_id"`    // 学院ID
    FacultyID    *int `form:"faculty_id"`    // 系ID
    ProfessionID *int `form:"profession_id"` // 专业ID（仅学生）
    ClassID      *int `form:"class_id"`      // 班级ID（仅学生）
}

// 人员基础信息
type Person struct {
    ID         int            `json:"id"`
    CustomerID int            `json:"customer_id"`
    PersonType int            `json:"person_type"`
    Name       string         `json:"name"`
    Gender     *int           `json:"gender"`
    Mobile     string         `json:"mobile"`
    Email      string         `json:"email"`
    Avatar     string         `json:"avatar"`
    Status     int            `json:"status"`
    
    // 扩展信息（根据 person_type 和 with_extend 动态填充）
    StudentExtend *StudentExtend `json:"student_extend,omitempty"`
    StaffExtend   *StaffExtend   `json:"staff_extend,omitempty"`
}

// 学生扩展信息
type StudentExtend struct {
    PersonID         int    `json:"person_id" gorm:"column:person_id"`
    AreaID           *int   `json:"area_id"`
    StudentNo        string `json:"student_no"`
    Grade            string `json:"grade"`
    EducationLevel   string `json:"education_level"`
    SchoolSystem     string `json:"school_system"`
    IDCard           string `json:"id_card"`
    AdmissionNo      string `json:"admission_no"`
    ExamNo           string `json:"exam_no"`
    EnrollmentStatus int    `json:"enrollment_status"`
    IsEnrolled       int    `json:"is_enrolled"`
    CollegeID        int    `json:"college_id"`
    FacultyID        *int   `json:"faculty_id"`
    ProfessionID     int    `json:"profession_id"`
    ClassID          *int   `json:"class_id"`
}

// 政工扩展信息
type StaffExtend struct {
    PersonID     int    `json:"person_id" gorm:"column:person_id"`
    StaffNo      string `json:"staff_no"`
    DepartmentID *int   `json:"department_id"`
    CollegeID    *int   `json:"college_id"`
    FacultyID    *int   `json:"faculty_id"`
}

// 人员列表响应
type PersonListResponse struct {
    Items    []Person `json:"items"`
    Total    int64    `json:"total"`
    Page     int      `json:"page"`
    PageSize int      `json:"page_size"`
}
```

## API 响应示例

### 学校管理员查询学生列表（不含扩展信息）

**请求：**
```http
GET /api/base/persons/list?university_id=1&person_type=1&page=1&page_size=20&name=张
Authorization: Bearer {admin_token}
```

**响应：**
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
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### 学校管理员查询学生列表（含扩展信息）

**请求：**
```http
GET /api/base/persons/list?university_id=1&person_type=1&page=1&page_size=20&with_extend=true&grade=2023
Authorization: Bearer {admin_token}
```

**响应：**
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
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### 政工人员查询政工列表（含扩展信息）

**请求：**
```http
GET /api/staff/persons/list?university_id=1&person_type=2&page=1&page_size=20&with_extend=true&department_id=5
Authorization: Bearer {staff_token}
```

**响应：**
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
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

## 缓存策略

### 政工权限缓存

**缓存时机：** 在政工人员登录时（`GET /api/staff/info`），直接计算并缓存 `managed_department_ids` 和 `managed_person_ids`

**缓存键：** `staff:managed_scope:{customer_id}:{person_id}`

**缓存内容：**
```json
{
  "managed_department_ids": [13, 20, 21, 46, 47, 48, 51, 52, 55, 61, 62, 63, 64],
  "managed_person_ids": [38, 37]
}
```

**说明：**
- `managed_department_ids`：管辖的所有部门ID（已扩展包含子部门）
- `managed_person_ids`：直接管辖的人员ID（不包含部门下的人员）
- 实际管辖的人员 = `managed_person_ids` + 管辖部门下的所有人员

```go
/**
 * 登录时缓存政工管辖权限
 * 在 GET /api/staff/info 接口中调用
 */
func CacheStaffManagedScopeOnLogin(customerId, personId int) error {
    // 1. 计算管辖的部门ID（包含子部门）
    managedDeptIds, managedPersonIds, err := GetStaffManagedScope(customerId, personId)
    if err != nil {
        return err
    }
    
    // 2. 构建缓存数据
    scope := struct {
        ManagedDepartmentIds []int `json:"managed_department_ids"`
        ManagedPersonIds     []int `json:"managed_person_ids"`
    }{
        ManagedDepartmentIds: managedDeptIds,
        ManagedPersonIds:     managedPersonIds,
    }
    
    // 3. 缓存到 Redis（永久有效，直到主动清除）
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    scopeJSON, _ := json.Marshal(scope)
    redis.Set(cacheKey, string(scopeJSON))
    
    return nil
}

/**
 * 获取政工管辖权限（从缓存读取）
 */
func GetStaffManagedScopeFromCache(customerId, personId int) ([]int, []int, error) {
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    
    // 从缓存获取
    cached, err := redis.Get(cacheKey)
    if err != nil || cached == "" {
        // 缓存不存在，重新计算并缓存
        return GetStaffManagedScopeAndCache(customerId, personId)
    }
    
    var scope struct {
        ManagedDepartmentIds []int `json:"managed_department_ids"`
        ManagedPersonIds     []int `json:"managed_person_ids"`
    }
    json.Unmarshal([]byte(cached), &scope)
    
    return scope.ManagedDepartmentIds, scope.ManagedPersonIds, nil
}

/**
 * 重新计算并缓存政工管辖权限
 */
func GetStaffManagedScopeAndCache(customerId, personId int) ([]int, []int, error) {
    // 计算管辖权限
    managedDeptIds, managedPersonIds, err := GetStaffManagedScope(customerId, personId)
    if err != nil {
        return nil, nil, err
    }
    
    // 缓存
    scope := struct {
        ManagedDepartmentIds []int `json:"managed_department_ids"`
        ManagedPersonIds     []int `json:"managed_person_ids"`
    }{
        ManagedDepartmentIds: managedDeptIds,
        ManagedPersonIds:     managedPersonIds,
    }
    
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    scopeJSON, _ := json.Marshal(scope)
    redis.Set(cacheKey, string(scopeJSON))
    
    return managedDeptIds, managedPersonIds, nil
}
```

### 缓存清除策略

**清除时机：** 当人员的部门信息发生变更时，清除该人员的管辖权限缓存

**触发场景：**
1. 人员的角色发生变化（`persons_has_roles` 表变更）
2. 角色的权限发生变化（`persons_roles` 表的 `department_ids` 或 `person_ids` 字段变更）
3. 人员的直接部门权限发生变化（`persons_has_department` 表变更）
4. 人员所属部门发生变化（`staff` 表的 `department_id`、`college_id`、`faculty_id` 字段变更）

```go
/**
 * 场景1：人员角色变更时，清除该人员的缓存
 * 触发时机：persons_has_roles 表 INSERT/UPDATE/DELETE
 */
func OnPersonRoleChanged(customerId, personId int) {
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    redis.Del(cacheKey)
}

/**
 * 场景2：角色权限变更时，清除拥有该角色的所有人员的缓存
 * 触发时机：persons_roles 表的 department_ids 或 person_ids 字段 UPDATE
 */
func OnRolePermissionChanged(customerId, roleId int) {
    // 查询拥有该角色的所有人员
    var personIds []int
    db.Table("persons_has_roles").
        Where("role_id = ?", roleId).
        Pluck("person_id", &personIds)
    
    // 清除这些人员的管辖权限缓存
    for _, personId := range personIds {
        cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
        redis.Del(cacheKey)
    }
}

/**
 * 场景3：人员直接部门权限变更时，清除该人员的缓存
 * 触发时机：persons_has_department 表 INSERT/UPDATE/DELETE
 */
func OnPersonDepartmentChanged(customerId, personId int) {
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    redis.Del(cacheKey)
}

/**
 * 场景4：人员所属部门变更时，清除该人员的缓存
 * 触发时机：staff 表的 department_id、college_id、faculty_id 字段 UPDATE
 */
func OnStaffDepartmentChanged(customerId, personId int) {
    cacheKey := fmt.Sprintf("staff:managed_scope:%d:%d", customerId, personId)
    redis.Del(cacheKey)
}

/**
 * 场景5：部门结构变更时，清除所有政工人员的缓存（可选）
 * 触发时机：departments 表的树结构发生变化
 * 注意：此场景影响范围较大，建议评估是否需要
 */
func OnDepartmentStructureChanged(customerId int) {
    // 方案1：清除该学校所有政工的缓存（简单但影响大）
    pattern := fmt.Sprintf("staff:managed_scope:%d:*", customerId)
    redis.DelPattern(pattern)
    
    // 方案2：只清除受影响部门相关的政工缓存（复杂但精确）
    // 需要查询哪些政工的管辖部门包含变更的部门
    // 然后逐个清除缓存
}
```

## 性能指标

| 接口类型 | P99 响应时间 | QPS | 缓存策略 |
|---------|-------------|-----|---------|
| 人员列表查询（不含扩展） | < 50ms | 500+ | 无缓存 |
| 人员列表查询（含扩展） | < 100ms | 300+ | 无缓存 |
| 政工权限查询 | < 10ms | 2000+ | Redis 永久（登录时缓存，变更时清除） |

## 索引优化建议

### persons 表索引

```sql
-- 复合索引：客户ID + 人员类型 + 删除标记
CREATE INDEX idx_customer_type_deleted ON persons(customer_id, person_type, deleted_at);

-- 姓名索引（支持模糊查询）
CREATE INDEX idx_name ON persons(name);

-- 手机号索引
CREATE INDEX idx_mobile ON persons(mobile);

-- 邮箱索引
CREATE INDEX idx_email ON persons(email);
```

### students 表索引

```sql
-- 人员ID索引（用于关联查询）
CREATE INDEX idx_person_id ON students(person_id);

-- 学号索引
CREATE UNIQUE INDEX uk_student_no ON students(student_no);

-- 组织结构索引
CREATE INDEX idx_college ON students(college_id);
CREATE INDEX idx_faculty ON students(faculty_id);
CREATE INDEX idx_profession ON students(profession_id);
CREATE INDEX idx_class ON students(class_id);

-- 年级索引
CREATE INDEX idx_grade ON students(grade);
```

### staff 表索引

```sql
-- 人员ID索引（用于关联查询）
CREATE INDEX idx_person_id ON staff(person_id);

-- 工号索引
CREATE UNIQUE INDEX uk_staff_no ON staff(staff_no);

-- 组织结构索引
CREATE INDEX idx_department ON staff(department_id);
CREATE INDEX idx_college ON staff(college_id);
CREATE INDEX idx_faculty ON staff(faculty_id);
```

## 注意事项

### 1. 权限过滤性能

政工人员查询时，权限过滤使用 `EXISTS` 子查询，需要确保：
- `students` 和 `staff` 表的 `person_id` 字段有索引
- 组织字段（college_id, faculty_id 等）有索引
- `managed_person_ids` 数组不宜过大（建议 < 1000）

### 2. 扩展信息查询优化

- 使用 `IN` 批量查询，避免 N+1 问题
- 每页人员数量建议控制在 20-50 之间
- 扩展信息查询使用 `person_id IN (...)` 确保使用索引

### 3. 搜索条件组合

- 支持多条件组合搜索
- 扩展字段搜索会影响性能，建议配合基础字段使用
- 复杂搜索场景可考虑使用 ElasticSearch

### 4. 分页限制

- 最大每页数量：100
- 深度分页（page > 100）性能较差，建议使用游标分页
- 总数统计在大数据量时可能较慢，可考虑缓存或估算

## 错误码定义

```go
const (
    ErrCodeSuccess           = 0     // 成功
    ErrCodeInvalidParam      = 1001  // 参数错误
    ErrCodeUnauthorized      = 1002  // 未授权
    ErrCodeForbidden         = 1003  // 无权限
    ErrCodeNotFound          = 1004  // 资源不存在
    ErrCodeInternalError     = 1005  // 内部错误
    ErrCodePersonTypeInvalid = 2001  // 人员类型无效
    ErrCodeNoPermission      = 2002  // 无查看权限
)
```

## 后续优化方向

1. **ElasticSearch 集成**
   - 将人员数据同步到 ES
   - 支持全文搜索和复杂条件查询
   - 提升大数据量下的查询性能

2. **读写分离**
   - 查询操作使用从库
   - 降低主库压力

3. **数据预热**
   - 常用查询条件的结果预缓存
   - 热点数据提前加载

4. **异步导出**
   - 大批量数据导出使用异步任务
   - 支持 Excel/CSV 格式导出
