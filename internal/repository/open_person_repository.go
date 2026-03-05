package repository

import (
	"college_api/internal/model"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// OpenPersonRepository 开放接口人员数据访问层
type OpenPersonRepository struct {
	db *sql.DB
}

// NewOpenPersonRepository 创建开放接口人员仓库
func NewOpenPersonRepository(db *sql.DB) *OpenPersonRepository {
	return &OpenPersonRepository{db: db}
}

// GetStaffList 查询政工列表
func (r *OpenPersonRepository) GetStaffList(req *model.OpenStaffRequest) (*model.OpenStaffResponse, error) {
	// 设置默认分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	// 构建基础查询
	baseFields := "s.person_id, COALESCE(s.name, '') as name, COALESCE(s.staff_no, '') as staff_no, s.department_id, s.college_id, s.faculty_id"
	if req.WithContact != nil && *req.WithContact == 1 {
		baseFields += ", p.mobile, p.email, p.gender, p.status"
	}

	query := fmt.Sprintf("SELECT %s FROM staff s", baseFields)
	if req.WithContact != nil && *req.WithContact == 1 {
		query += " LEFT JOIN persons p ON s.person_id = p.id"
	}
	query += " WHERE s.university_id = ?"
	args := []any{req.UniversityID}

	// 添加过滤条件
	query, args = r.addStaffFilters(query, args, req)

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 游标分页
	query += " ORDER BY s.person_id ASC LIMIT ? OFFSET ?"
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenStaffItem
	for rows.Next() {
		var item model.OpenStaffItem
		var name, staffNo sql.NullString
		var deptID, collegeID, facultyID sql.NullInt64

		if req.WithContact != nil && *req.WithContact == 1 {
			var mobile, email sql.NullString
			var gender, status sql.NullInt64
			err = rows.Scan(&item.PersonID, &name, &staffNo, &deptID, &collegeID, &facultyID,
				&mobile, &email, &gender, &status)
			if err != nil {
				return nil, err
			}
			if mobile.Valid {
				item.Mobile = &mobile.String
			}
			if email.Valid {
				item.Email = &email.String
			}
			if gender.Valid {
				g := int(gender.Int64)
				item.Gender = &g
			}
			if status.Valid {
				s := int(status.Int64)
				item.Status = &s
			}
		} else {
			err = rows.Scan(&item.PersonID, &name, &staffNo, &deptID, &collegeID, &facultyID)
			if err != nil {
				return nil, err
			}
		}

		item.Name = name.String
		item.StaffNo = staffNo.String
		if deptID.Valid {
			d := int(deptID.Int64)
			item.DepartmentID = &d
		}
		if collegeID.Valid {
			c := int(collegeID.Int64)
			item.CollegeID = &c
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			item.FacultyID = &f
		}

		items = append(items, item)
	}

	return &model.OpenStaffResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (r *OpenPersonRepository) addStaffFilters(query string, args []any, req *model.OpenStaffRequest) (string, []any) {
	// person_id 支持逗号分隔
	if req.PersonIDs != "" {
		ids := r.parseIntIDs(req.PersonIDs)
		if len(ids) > 0 {
			placeholders := make([]string, len(ids))
			for i, id := range ids {
				placeholders[i] = "?"
				args = append(args, id)
			}
			query += fmt.Sprintf(" AND s.person_id IN (%s)", strings.Join(placeholders, ","))
		}
	}

	// name 支持逗号分隔
	if req.Names != "" {
		names := strings.Split(req.Names, ",")
		if len(names) > 0 {
			conditions := make([]string, 0, len(names))
			for _, name := range names {
				name = strings.TrimSpace(name)
				if name != "" {
					conditions = append(conditions, "s.name LIKE ?")
					args = append(args, "%"+name+"%")
				}
			}
			if len(conditions) > 0 {
				query += " AND (" + strings.Join(conditions, " OR ") + ")"
			}
		}
	}

	// staff_no 支持逗号分隔
	if req.StaffNos != "" {
		nos := strings.Split(req.StaffNos, ",")
		if len(nos) > 0 {
			placeholders := make([]string, 0, len(nos))
			for _, no := range nos {
				no = strings.TrimSpace(no)
				if no != "" {
					placeholders = append(placeholders, "?")
					args = append(args, no)
				}
			}
			if len(placeholders) > 0 {
				query += fmt.Sprintf(" AND s.staff_no IN (%s)", strings.Join(placeholders, ","))
			}
		}
	}

	if req.DepartmentIDs != "" {
		query, args = r.addInFilter(query, args, "s.department_id", req.DepartmentIDs)
	} else if req.DepartmentID != nil {
		query += " AND s.department_id = ?"
		args = append(args, *req.DepartmentID)
	}
	if req.CollegeIDs != "" {
		query, args = r.addInFilter(query, args, "s.college_id", req.CollegeIDs)
	} else if req.CollegeID != nil {
		query += " AND s.college_id = ?"
		args = append(args, *req.CollegeID)
	}
	if req.FacultyIDs != "" {
		query, args = r.addInFilter(query, args, "s.faculty_id", req.FacultyIDs)
	} else if req.FacultyID != nil {
		query += " AND s.faculty_id = ?"
		args = append(args, *req.FacultyID)
	}

	return query, args
}

// GetStudentList 查询学生列表
func (r *OpenPersonRepository) GetStudentList(req *model.OpenStudentRequest) (*model.OpenStudentResponse, error) {
	// 设置默认分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	// 构建基础查询
	baseFields := `s.person_id, COALESCE(s.name, '') as name, COALESCE(s.student_no, '') as student_no,
		s.area_id, COALESCE(s.grade, '') as grade, COALESCE(s.education_level, '') as education_level,
		COALESCE(s.school_system, '') as school_system, COALESCE(s.id_card, '') as id_card,
		COALESCE(s.admission_no, '') as admission_no, COALESCE(s.exam_no, '') as exam_no,
		s.enrollment_status, s.is_enrolled, s.college_id, s.faculty_id, s.profession_id, s.class_id`
	if req.WithContact != nil && *req.WithContact == 1 {
		baseFields += ", p.mobile, p.email, p.gender, p.status"
	}

	query := fmt.Sprintf("SELECT %s FROM students s", baseFields)
	if req.WithContact != nil && *req.WithContact == 1 {
		query += " LEFT JOIN persons p ON s.person_id = p.id"
	}
	query += " WHERE s.university_id = ?"
	args := []any{req.UniversityID}

	// 添加过滤条件
	query, args = r.addStudentFilters(query, args, req)

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 游标分页
	query += " ORDER BY s.person_id ASC LIMIT ? OFFSET ?"
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenStudentItem
	for rows.Next() {
		var item model.OpenStudentItem
		var name, studentNo, grade, educationLevel, schoolSystem, idCard, admissionNo, examNo sql.NullString
		var areaID, facultyID, classID sql.NullInt64

		if req.WithContact != nil && *req.WithContact == 1 {
			var mobile, email sql.NullString
			var gender, status sql.NullInt64
			err = rows.Scan(&item.PersonID, &name, &studentNo, &areaID, &grade, &educationLevel,
				&schoolSystem, &idCard, &admissionNo, &examNo, &item.EnrollmentStatus, &item.IsEnrolled,
				&item.CollegeID, &facultyID, &item.ProfessionID, &classID,
				&mobile, &email, &gender, &status)
			if err != nil {
				return nil, err
			}
			if mobile.Valid {
				item.Mobile = &mobile.String
			}
			if email.Valid {
				item.Email = &email.String
			}
			if gender.Valid {
				g := int(gender.Int64)
				item.Gender = &g
			}
			if status.Valid {
				s := int(status.Int64)
				item.Status = &s
			}
		} else {
			err = rows.Scan(&item.PersonID, &name, &studentNo, &areaID, &grade, &educationLevel,
				&schoolSystem, &idCard, &admissionNo, &examNo, &item.EnrollmentStatus, &item.IsEnrolled,
				&item.CollegeID, &facultyID, &item.ProfessionID, &classID)
			if err != nil {
				return nil, err
			}
		}

		item.Name = name.String
		item.StudentNo = studentNo.String
		item.Grade = grade.String
		item.EducationLevel = educationLevel.String
		item.SchoolSystem = schoolSystem.String
		item.IDCard = idCard.String
		item.AdmissionNo = admissionNo.String
		item.ExamNo = examNo.String
		if areaID.Valid {
			a := int(areaID.Int64)
			item.AreaID = &a
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			item.FacultyID = &f
		}
		if classID.Valid {
			c := int(classID.Int64)
			item.ClassID = &c
		}

		items = append(items, item)
	}

	return &model.OpenStudentResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (r *OpenPersonRepository) addStudentFilters(query string, args []any, req *model.OpenStudentRequest) (string, []any) {
	// person_id 支持逗号分隔
	if req.PersonIDs != "" {
		ids := r.parseIntIDs(req.PersonIDs)
		if len(ids) > 0 {
			placeholders := make([]string, len(ids))
			for i, id := range ids {
				placeholders[i] = "?"
				args = append(args, id)
			}
			query += fmt.Sprintf(" AND s.person_id IN (%s)", strings.Join(placeholders, ","))
		}
	}

	// name 支持逗号分隔
	if req.Names != "" {
		names := strings.Split(req.Names, ",")
		if len(names) > 0 {
			conditions := make([]string, 0, len(names))
			for _, name := range names {
				name = strings.TrimSpace(name)
				if name != "" {
					conditions = append(conditions, "s.name LIKE ?")
					args = append(args, "%"+name+"%")
				}
			}
			if len(conditions) > 0 {
				query += " AND (" + strings.Join(conditions, " OR ") + ")"
			}
		}
	}

	// student_no 支持逗号分隔
	if req.StudentNos != "" {
		nos := strings.Split(req.StudentNos, ",")
		if len(nos) > 0 {
			placeholders := make([]string, 0, len(nos))
			for _, no := range nos {
				no = strings.TrimSpace(no)
				if no != "" {
					placeholders = append(placeholders, "?")
					args = append(args, no)
				}
			}
			if len(placeholders) > 0 {
				query += fmt.Sprintf(" AND s.student_no IN (%s)", strings.Join(placeholders, ","))
			}
		}
	}

	if req.AreaID != nil {
		query += " AND s.area_id = ?"
		args = append(args, *req.AreaID)
	}
	if req.Grade != "" {
		query += " AND s.grade = ?"
		args = append(args, req.Grade)
	}
	if req.EducationLevel != "" {
		query += " AND s.education_level = ?"
		args = append(args, req.EducationLevel)
	}
	if req.SchoolSystem != "" {
		query += " AND s.school_system = ?"
		args = append(args, req.SchoolSystem)
	}
	if req.IDCard != "" {
		query += " AND s.id_card = ?"
		args = append(args, req.IDCard)
	}
	if req.AdmissionNo != "" {
		query += " AND s.admission_no = ?"
		args = append(args, req.AdmissionNo)
	}
	if req.ExamNo != "" {
		query += " AND s.exam_no = ?"
		args = append(args, req.ExamNo)
	}
	if req.EnrollmentStatus != nil {
		query += " AND s.enrollment_status = ?"
		args = append(args, *req.EnrollmentStatus)
	}
	if req.IsEnrolled != nil {
		query += " AND s.is_enrolled = ?"
		args = append(args, *req.IsEnrolled)
	}
	if req.CollegeIDs != "" {
		query, args = r.addInFilter(query, args, "s.college_id", req.CollegeIDs)
	} else if req.CollegeID != nil {
		query += " AND s.college_id = ?"
		args = append(args, *req.CollegeID)
	}
	if req.FacultyIDs != "" {
		query, args = r.addInFilter(query, args, "s.faculty_id", req.FacultyIDs)
	} else if req.FacultyID != nil {
		query += " AND s.faculty_id = ?"
		args = append(args, *req.FacultyID)
	}
	if req.ProfessionIDs != "" {
		query, args = r.addInFilter(query, args, "s.profession_id", req.ProfessionIDs)
	} else if req.ProfessionID != nil {
		query += " AND s.profession_id = ?"
		args = append(args, *req.ProfessionID)
	}
	if req.ClassIDs != "" {
		query, args = r.addInFilter(query, args, "s.class_id", req.ClassIDs)
	} else if req.ClassID != nil {
		query += " AND s.class_id = ?"
		args = append(args, *req.ClassID)
	}

	return query, args
}

// GetManagePersons 查询管辖某人员的所有管理者（含角色和管辖机构）
func (r *OpenPersonRepository) GetManagePersons(universityID, personID int) ([]model.OpenManagePersonItem, error) {
	// 1. 查询目标人员的类型
	var personType int
	err := r.db.QueryRow(
		"SELECT person_type FROM persons WHERE id = ? AND customer_id = ? AND deleted_at = 0 LIMIT 1",
		personID, universityID,
	).Scan(&personType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	// 2. 获取目标人员所属的机构ID列表
	deptIDs, err := r.getPersonDepartmentIDs(personID, personType)
	if err != nil {
		return nil, err
	}
	if len(deptIDs) == 0 {
		return []model.OpenManagePersonItem{}, nil
	}

	// 3. 从 persons_has_department 查找管辖这些机构的人员（排除自己）
	// 管辖配置时选择上级机构会自动勾选下级，所以表中已包含所有被管辖的机构ID，无需查祖先
	query := `
		SELECT DISTINCT phd.person_id, phd.persons_roles_id, phd.department_id
		FROM persons_has_department phd
		WHERE phd.customer_id = ? AND phd.person_id != ?
		  AND phd.department_id IN (` + r.buildPlaceholders(len(deptIDs)) + `)
		ORDER BY phd.person_id, phd.persons_roles_id
	`
	args := []any{universityID, personID}
	for _, id := range deptIDs {
		args = append(args, id)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query manage persons error: %w", err)
	}
	defer rows.Close()

	// personID -> roleID -> []departmentID
	type roleDept struct {
		roleID int
		deptID int
	}
	personRoleDepts := make(map[int][]roleDept)
	var personIDs []int
	personIDSet := make(map[int]bool)

	for rows.Next() {
		var pID, roleID, deptID int
		if err := rows.Scan(&pID, &roleID, &deptID); err != nil {
			return nil, err
		}
		personRoleDepts[pID] = append(personRoleDepts[pID], roleDept{roleID: roleID, deptID: deptID})
		if !personIDSet[pID] {
			personIDSet[pID] = true
			personIDs = append(personIDs, pID)
		}
	}

	if len(personIDs) == 0 {
		return []model.OpenManagePersonItem{}, nil
	}

	// 5. 批量查询管理者姓名
	personNames, err := r.batchGetPersonNames(personIDs)
	if err != nil {
		return nil, err
	}

	// 6. 收集所有涉及的角色ID和机构ID，批量查询名称
	roleIDSet := make(map[int]bool)
	deptIDSet := make(map[int]bool)
	for _, rds := range personRoleDepts {
		for _, rd := range rds {
			roleIDSet[rd.roleID] = true
			deptIDSet[rd.deptID] = true
		}
	}

	roleNames, err := r.batchGetRoleNames(universityID, roleIDSet)
	if err != nil {
		return nil, err
	}

	deptNames, err := r.batchGetDepartmentNames(deptIDSet)
	if err != nil {
		return nil, err
	}

	// 7. 组装结果
	var result []model.OpenManagePersonItem
	for _, pID := range personIDs {
		rds := personRoleDepts[pID]

		// 按角色分组
		roleMap := make(map[int][]model.OpenManagePersonDepartment)
		var roleOrder []int
		roleOrderSet := make(map[int]bool)
		for _, rd := range rds {
			if !roleOrderSet[rd.roleID] {
				roleOrderSet[rd.roleID] = true
				roleOrder = append(roleOrder, rd.roleID)
			}
			roleMap[rd.roleID] = append(roleMap[rd.roleID], model.OpenManagePersonDepartment{
				DepartmentID:   rd.deptID,
				DepartmentName: deptNames[rd.deptID],
			})
		}

		var roles []model.OpenManagePersonRole
		for _, rID := range roleOrder {
			roles = append(roles, model.OpenManagePersonRole{
				RoleID:      rID,
				RoleName:    roleNames[rID],
				Departments: roleMap[rID],
			})
		}

		result = append(result, model.OpenManagePersonItem{
			PersonID:   pID,
			PersonName: personNames[pID],
			Roles:      roles,
		})
	}

	return result, nil
}

// getPersonDepartmentIDs 获取人员所属的机构ID列表
func (r *OpenPersonRepository) getPersonDepartmentIDs(personID, personType int) ([]int, error) {
	var deptIDs []int
	deptSet := make(map[int]bool)

	addDept := func(id int) {
		if id > 0 && !deptSet[id] {
			deptSet[id] = true
			deptIDs = append(deptIDs, id)
		}
	}

	if personType == 1 { // 学生
		var collegeID, professionID sql.NullInt64
		var facultyID, classID sql.NullInt64
		err := r.db.QueryRow(
			"SELECT college_id, faculty_id, profession_id, class_id FROM students WHERE person_id = ? LIMIT 1",
			personID,
		).Scan(&collegeID, &facultyID, &professionID, &classID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		if collegeID.Valid {
			addDept(int(collegeID.Int64))
		}
		if facultyID.Valid {
			addDept(int(facultyID.Int64))
		}
		if professionID.Valid {
			addDept(int(professionID.Int64))
		}
		if classID.Valid {
			addDept(int(classID.Int64))
		}
	} else { // 政工/维修工
		var departmentID, collegeID, facultyID sql.NullInt64
		err := r.db.QueryRow(
			"SELECT department_id, college_id, faculty_id FROM staff WHERE person_id = ? LIMIT 1",
			personID,
		).Scan(&departmentID, &collegeID, &facultyID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		if departmentID.Valid {
			addDept(int(departmentID.Int64))
		}
		if collegeID.Valid {
			addDept(int(collegeID.Int64))
		}
		if facultyID.Valid {
			addDept(int(facultyID.Int64))
		}
	}

	return deptIDs, nil
}

// batchGetPersonNames 批量获取人员姓名
func (r *OpenPersonRepository) batchGetPersonNames(personIDs []int) (map[int]string, error) {
	if len(personIDs) == 0 {
		return make(map[int]string), nil
	}

	query := "SELECT id, name FROM persons WHERE id IN (" + r.buildPlaceholders(len(personIDs)) + ") AND deleted_at = 0"
	args := make([]any, len(personIDs))
	for i, id := range personIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[int]string)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		names[id] = name
	}
	return names, nil
}

// batchGetRoleNames 批量获取角色名称
func (r *OpenPersonRepository) batchGetRoleNames(universityID int, roleIDSet map[int]bool) (map[int]string, error) {
	if len(roleIDSet) == 0 {
		return make(map[int]string), nil
	}

	ids := make([]int, 0, len(roleIDSet))
	for id := range roleIDSet {
		ids = append(ids, id)
	}

	query := "SELECT id, name FROM persons_roles WHERE customer_id = ? AND id IN (" + r.buildPlaceholders(len(ids)) + ") AND deleted_at = 0"
	args := []any{universityID}
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[int]string)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		names[id] = name
	}
	return names, nil
}

// batchGetDepartmentNames 批量获取机构名称
func (r *OpenPersonRepository) batchGetDepartmentNames(deptIDSet map[int]bool) (map[int]string, error) {
	if len(deptIDSet) == 0 {
		return make(map[int]string), nil
	}

	ids := make([]int, 0, len(deptIDSet))
	for id := range deptIDSet {
		ids = append(ids, id)
	}

	query := "SELECT id, department_name FROM departments WHERE id IN (" + r.buildPlaceholders(len(ids)) + ") AND deleted_at = 0"
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[int]string)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		names[id] = name
	}
	return names, nil
}

// buildPlaceholders 构建SQL占位符
func (r *OpenPersonRepository) buildPlaceholders(count int) string {
	if count == 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

func (r *OpenPersonRepository) parseIntIDs(idsStr string) []int {
	if idsStr == "" {
		return nil
	}
	var result []int
	for _, s := range strings.Split(idsStr, ",") {
		s = strings.TrimSpace(s)
		if id, err := strconv.Atoi(s); err == nil {
			result = append(result, id)
		}
	}
	return result
}

// addInFilter 通用的多ID IN查询条件构建
func (r *OpenPersonRepository) addInFilter(query string, args []any, column string, idsStr string) (string, []any) {
	ids := r.parseIntIDs(idsStr)
	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND %s IN (%s)", column, strings.Join(placeholders, ","))
	}
	return query, args
}

// GetRolePersons 根据角色查询人员详情（支持多角色、分页）
func (r *OpenPersonRepository) GetRolePersons(req *model.OpenRolePersonsRequest) (*model.OpenRolePersonsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	// 解析角色ID
	roleIDs, err := r.resolveRoleIDs(req)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return &model.OpenRolePersonsResponse{Items: []model.OpenRolePersonItem{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil
	}

	// 查询去重后的人员总数
	countQuery := "SELECT COUNT(DISTINCT person_id) FROM persons_has_roles WHERE customer_id = ? AND role_id IN (" + r.buildPlaceholders(len(roleIDs)) + ")"
	countArgs := []any{req.UniversityID}
	for _, id := range roleIDs {
		countArgs = append(countArgs, id)
	}
	var total int64
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	if total == 0 {
		return &model.OpenRolePersonsResponse{Items: []model.OpenRolePersonItem{}, Total: 0, Page: req.Page, PageSize: req.PageSize}, nil
	}

	// 分页查询人员ID
	pidQuery := "SELECT DISTINCT person_id FROM persons_has_roles WHERE customer_id = ? AND role_id IN (" + r.buildPlaceholders(len(roleIDs)) + ") ORDER BY person_id ASC LIMIT ? OFFSET ?"
	pidArgs := []any{req.UniversityID}
	for _, id := range roleIDs {
		pidArgs = append(pidArgs, id)
	}
	offset := (req.Page - 1) * req.PageSize
	pidArgs = append(pidArgs, req.PageSize, offset)

	pidRows, err := r.db.Query(pidQuery, pidArgs...)
	if err != nil {
		return nil, fmt.Errorf("query person_ids error: %w", err)
	}
	defer pidRows.Close()

	var personIDs []int
	for pidRows.Next() {
		var pid int
		if err := pidRows.Scan(&pid); err != nil {
			return nil, err
		}
		personIDs = append(personIDs, pid)
	}

	if len(personIDs) == 0 {
		return &model.OpenRolePersonsResponse{Items: []model.OpenRolePersonItem{}, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
	}

	// 批量查询人员基础信息
	items, err := r.batchGetPersonBaseInfo(req.UniversityID, personIDs)
	if err != nil {
		return nil, err
	}

	// 按人员类型分组，补充扩展信息
	var studentPIDs, staffPIDs []int
	itemMap := make(map[int]*model.OpenRolePersonItem)
	for i := range items {
		itemMap[items[i].PersonID] = &items[i]
		if items[i].PersonType == 1 {
			studentPIDs = append(studentPIDs, items[i].PersonID)
		} else {
			staffPIDs = append(staffPIDs, items[i].PersonID)
		}
	}

	// 补充学生扩展信息
	if len(studentPIDs) > 0 {
		if err := r.enrichStudentInfo(req.UniversityID, studentPIDs, itemMap); err != nil {
			return nil, err
		}
	}

	// 补充政工扩展信息
	if len(staffPIDs) > 0 {
		if err := r.enrichStaffInfo(req.UniversityID, staffPIDs, itemMap); err != nil {
			return nil, err
		}
	}

	// 批量查询机构名称并填充
	deptIDs := r.collectDeptIDs(items)
	deptNames, err := r.batchGetDepartmentNames(deptIDs)
	if err != nil {
		return nil, err
	}
	r.fillDeptNames(items, deptNames)

	return &model.OpenRolePersonsResponse{Items: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// resolveRoleIDs 解析请求中的角色ID（支持 role_ids 或 role_name）
func (r *OpenPersonRepository) resolveRoleIDs(req *model.OpenRolePersonsRequest) ([]int, error) {
	if req.RoleIDs != "" {
		return r.parseIntIDs(req.RoleIDs), nil
	}
	if req.RoleName != "" {
		var roleID int
		err := r.db.QueryRow(
			"SELECT id FROM persons_roles WHERE customer_id = ? AND name = ? AND deleted_at = 0 LIMIT 1",
			req.UniversityID, req.RoleName,
		).Scan(&roleID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, fmt.Errorf("query role error: %w", err)
		}
		return []int{roleID}, nil
	}
	return nil, fmt.Errorf("role_ids or role_name is required")
}

// batchGetPersonBaseInfo 批量查询人员基础信息
func (r *OpenPersonRepository) batchGetPersonBaseInfo(universityID int, personIDs []int) ([]model.OpenRolePersonItem, error) {
	query := "SELECT id, person_type, name, mobile, status FROM persons WHERE customer_id = ? AND id IN (" + r.buildPlaceholders(len(personIDs)) + ") AND deleted_at = 0 ORDER BY id ASC"
	args := []any{universityID}
	for _, id := range personIDs {
		args = append(args, id)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query persons error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenRolePersonItem
	for rows.Next() {
		var item model.OpenRolePersonItem
		var mobile sql.NullString
		if err := rows.Scan(&item.PersonID, &item.PersonType, &item.Name, &mobile, &item.Status); err != nil {
			return nil, err
		}
		if mobile.Valid {
			item.Mobile = &mobile.String
		}
		items = append(items, item)
	}
	return items, nil
}

// enrichStudentInfo 补充学生扩展信息（学号、组织机构）
func (r *OpenPersonRepository) enrichStudentInfo(universityID int, personIDs []int, itemMap map[int]*model.OpenRolePersonItem) error {
	query := "SELECT person_id, COALESCE(student_no, '') as student_no, college_id, faculty_id, profession_id, class_id FROM students WHERE university_id = ? AND person_id IN (" + r.buildPlaceholders(len(personIDs)) + ")"
	args := []any{universityID}
	for _, id := range personIDs {
		args = append(args, id)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("query students error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pid int
		var studentNo string
		var collegeID, professionID sql.NullInt64
		var facultyID, classID sql.NullInt64
		if err := rows.Scan(&pid, &studentNo, &collegeID, &facultyID, &professionID, &classID); err != nil {
			return err
		}
		item, ok := itemMap[pid]
		if !ok {
			continue
		}
		item.StudentNo = &studentNo
		if collegeID.Valid {
			c := int(collegeID.Int64)
			item.CollegeID = &c
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			item.FacultyID = &f
		}
		if professionID.Valid {
			p := int(professionID.Int64)
			item.ProfessionID = &p
		}
		if classID.Valid {
			cl := int(classID.Int64)
			item.ClassID = &cl
		}
	}
	return nil
}

// enrichStaffInfo 补充政工扩展信息（工号、组织机构）
func (r *OpenPersonRepository) enrichStaffInfo(universityID int, personIDs []int, itemMap map[int]*model.OpenRolePersonItem) error {
	query := "SELECT person_id, COALESCE(staff_no, '') as staff_no, department_id, college_id, faculty_id FROM staff WHERE university_id = ? AND person_id IN (" + r.buildPlaceholders(len(personIDs)) + ")"
	args := []any{universityID}
	for _, id := range personIDs {
		args = append(args, id)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("query staff error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pid int
		var staffNo string
		var deptID, collegeID, facultyID sql.NullInt64
		if err := rows.Scan(&pid, &staffNo, &deptID, &collegeID, &facultyID); err != nil {
			return err
		}
		item, ok := itemMap[pid]
		if !ok {
			continue
		}
		item.StaffNo = &staffNo
		if deptID.Valid {
			d := int(deptID.Int64)
			item.DepartmentID = &d
		}
		if collegeID.Valid {
			c := int(collegeID.Int64)
			item.CollegeID = &c
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			item.FacultyID = &f
		}
	}
	return nil
}

// collectDeptIDs 收集所有机构ID用于批量查询名称
func (r *OpenPersonRepository) collectDeptIDs(items []model.OpenRolePersonItem) map[int]bool {
	ids := make(map[int]bool)
	for _, item := range items {
		if item.CollegeID != nil {
			ids[*item.CollegeID] = true
		}
		if item.FacultyID != nil {
			ids[*item.FacultyID] = true
		}
		if item.DepartmentID != nil {
			ids[*item.DepartmentID] = true
		}
		if item.ProfessionID != nil {
			ids[*item.ProfessionID] = true
		}
		if item.ClassID != nil {
			ids[*item.ClassID] = true
		}
	}
	return ids
}

// fillDeptNames 将机构名称填充到人员详情中
func (r *OpenPersonRepository) fillDeptNames(items []model.OpenRolePersonItem, names map[int]string) {
	for i := range items {
		if items[i].CollegeID != nil {
			if n, ok := names[*items[i].CollegeID]; ok {
				items[i].CollegeName = &n
			}
		}
		if items[i].FacultyID != nil {
			if n, ok := names[*items[i].FacultyID]; ok {
				items[i].FacultyName = &n
			}
		}
		if items[i].DepartmentID != nil {
			if n, ok := names[*items[i].DepartmentID]; ok {
				items[i].DepartmentName = &n
			}
		}
		if items[i].ProfessionID != nil {
			if n, ok := names[*items[i].ProfessionID]; ok {
				items[i].ProfessionName = &n
			}
		}
		if items[i].ClassID != nil {
			if n, ok := names[*items[i].ClassID]; ok {
				items[i].ClassName = &n
			}
		}
	}
}

// GetRoleList 查询角色列表
func (r *OpenPersonRepository) GetRoleList(req *model.OpenRoleRequest) ([]model.OpenRoleItem, error) {
	query := `SELECT id, customer_id, parent_id, name, COALESCE(permissions, '') as permissions 
		FROM persons_roles WHERE customer_id = ? AND deleted_at = 0`

	rows, err := r.db.Query(query, req.UniversityID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenRoleItem
	for rows.Next() {
		var item model.OpenRoleItem
		var customerID int
		if err := rows.Scan(&item.ID, &customerID, &item.ParentID, &item.Name, &item.Permissions); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetStaffByOrg 按组织机构查询政工（OR条件）
// SQL: SELECT * FROM staff WHERE university_id = ? AND (department_id IN (?) OR college_id IN (?) OR faculty_id IN (?))
func (r *OpenPersonRepository) GetStaffByOrg(req *model.OpenStaffByOrgRequest) (*model.OpenStaffResponse, error) {
	// 设置默认分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	// 解析ID列表
	deptIDs := r.parseIntIDs(req.DepartmentIDs)
	collegeIDs := r.parseIntIDs(req.CollegeIDs)
	facultyIDs := r.parseIntIDs(req.FacultyIDs)

	// 至少需要一个组织条件
	if len(deptIDs) == 0 && len(collegeIDs) == 0 && len(facultyIDs) == 0 {
		return &model.OpenStaffResponse{
			Items:    []model.OpenStaffItem{},
			Total:    0,
			Page:     req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	// 构建基础查询
	baseFields := "s.person_id, COALESCE(s.name, '') as name, COALESCE(s.staff_no, '') as staff_no, s.department_id, s.college_id, s.faculty_id"
	if req.WithContact != nil && *req.WithContact == 1 {
		baseFields += ", p.mobile, p.email, p.gender, p.status"
	}

	query := fmt.Sprintf("SELECT %s FROM staff s", baseFields)
	if req.WithContact != nil && *req.WithContact == 1 {
		query += " LEFT JOIN persons p ON s.person_id = p.id"
	}
	query += " WHERE s.university_id = ?"
	args := []any{req.UniversityID}

	// 构建OR条件
	var orConditions []string
	if len(deptIDs) > 0 {
		placeholders := make([]string, len(deptIDs))
		for i, id := range deptIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.department_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(collegeIDs) > 0 {
		placeholders := make([]string, len(collegeIDs))
		for i, id := range collegeIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.college_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(facultyIDs) > 0 {
		placeholders := make([]string, len(facultyIDs))
		for i, id := range facultyIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.faculty_id IN (%s)", strings.Join(placeholders, ",")))
	}

	query += " AND (" + strings.Join(orConditions, " OR ") + ")"

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 分页
	query += " ORDER BY s.person_id ASC LIMIT ? OFFSET ?"
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenStaffItem
	for rows.Next() {
		var item model.OpenStaffItem
		var name, staffNo sql.NullString
		var deptID, collegeID, facultyID sql.NullInt64

		if req.WithContact != nil && *req.WithContact == 1 {
			var mobile, email sql.NullString
			var gender, status sql.NullInt64
			err = rows.Scan(&item.PersonID, &name, &staffNo, &deptID, &collegeID, &facultyID,
				&mobile, &email, &gender, &status)
			if err != nil {
				return nil, err
			}
			if mobile.Valid {
				item.Mobile = &mobile.String
			}
			if email.Valid {
				item.Email = &email.String
			}
			if gender.Valid {
				g := int(gender.Int64)
				item.Gender = &g
			}
			if status.Valid {
				s := int(status.Int64)
				item.Status = &s
			}
		} else {
			err = rows.Scan(&item.PersonID, &name, &staffNo, &deptID, &collegeID, &facultyID)
			if err != nil {
				return nil, err
			}
		}

		item.Name = name.String
		item.StaffNo = staffNo.String
		if deptID.Valid {
			d := int(deptID.Int64)
			item.DepartmentID = &d
		}
		if collegeID.Valid {
			c := int(collegeID.Int64)
			item.CollegeID = &c
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			item.FacultyID = &f
		}

		items = append(items, item)
	}

	return &model.OpenStaffResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetStudentByOrg 按组织机构查询学生（OR条件）
// SQL: SELECT * FROM students WHERE university_id = ? AND (college_id IN (?) OR faculty_id IN (?) OR profession_id IN (?) OR class_id IN (?))
func (r *OpenPersonRepository) GetStudentByOrg(req *model.OpenStudentByOrgRequest) (*model.OpenStudentResponse, error) {
	// 设置默认分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	// 解析ID列表
	collegeIDs := r.parseIntIDs(req.CollegeIDs)
	facultyIDs := r.parseIntIDs(req.FacultyIDs)
	professionIDs := r.parseIntIDs(req.ProfessionIDs)
	classIDs := r.parseIntIDs(req.ClassIDs)

	// 至少需要一个组织条件
	if len(collegeIDs) == 0 && len(facultyIDs) == 0 && len(professionIDs) == 0 && len(classIDs) == 0 {
		return &model.OpenStudentResponse{
			Items:    []model.OpenStudentItem{},
			Total:    0,
			Page:     req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	// 构建基础查询
	baseFields := `s.person_id, COALESCE(s.name, '') as name, COALESCE(s.student_no, '') as student_no,
		s.area_id, COALESCE(s.grade, '') as grade, COALESCE(s.education_level, '') as education_level,
		COALESCE(s.school_system, '') as school_system, COALESCE(s.id_card, '') as id_card,
		COALESCE(s.admission_no, '') as admission_no, COALESCE(s.exam_no, '') as exam_no,
		s.enrollment_status, s.is_enrolled, s.college_id, s.faculty_id, s.profession_id, s.class_id`
	if req.WithContact != nil && *req.WithContact == 1 {
		baseFields += ", p.mobile, p.email, p.gender, p.status"
	}

	query := fmt.Sprintf("SELECT %s FROM students s", baseFields)
	if req.WithContact != nil && *req.WithContact == 1 {
		query += " LEFT JOIN persons p ON s.person_id = p.id"
	}
	query += " WHERE s.university_id = ?"
	args := []any{req.UniversityID}

	// 构建OR条件
	var orConditions []string
	if len(collegeIDs) > 0 {
		placeholders := make([]string, len(collegeIDs))
		for i, id := range collegeIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.college_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(facultyIDs) > 0 {
		placeholders := make([]string, len(facultyIDs))
		for i, id := range facultyIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.faculty_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(professionIDs) > 0 {
		placeholders := make([]string, len(professionIDs))
		for i, id := range professionIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.profession_id IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(classIDs) > 0 {
		placeholders := make([]string, len(classIDs))
		for i, id := range classIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		orConditions = append(orConditions, fmt.Sprintf("s.class_id IN (%s)", strings.Join(placeholders, ",")))
	}

	query += " AND (" + strings.Join(orConditions, " OR ") + ")"

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 分页
	query += " ORDER BY s.person_id ASC LIMIT ? OFFSET ?"
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenStudentItem
	for rows.Next() {
		var item model.OpenStudentItem
		var name, studentNo, grade, educationLevel, schoolSystem, idCard, admissionNo, examNo sql.NullString
		var areaID, facultyID, classID sql.NullInt64

		if req.WithContact != nil && *req.WithContact == 1 {
			var mobile, email sql.NullString
			var gender, status sql.NullInt64
			err = rows.Scan(&item.PersonID, &name, &studentNo, &areaID, &grade, &educationLevel,
				&schoolSystem, &idCard, &admissionNo, &examNo, &item.EnrollmentStatus, &item.IsEnrolled,
				&item.CollegeID, &facultyID, &item.ProfessionID, &classID,
				&mobile, &email, &gender, &status)
			if err != nil {
				return nil, err
			}
			if mobile.Valid {
				item.Mobile = &mobile.String
			}
			if email.Valid {
				item.Email = &email.String
			}
			if gender.Valid {
				g := int(gender.Int64)
				item.Gender = &g
			}
			if status.Valid {
				s := int(status.Int64)
				item.Status = &s
			}
		} else {
			err = rows.Scan(&item.PersonID, &name, &studentNo, &areaID, &grade, &educationLevel,
				&schoolSystem, &idCard, &admissionNo, &examNo, &item.EnrollmentStatus, &item.IsEnrolled,
				&item.CollegeID, &facultyID, &item.ProfessionID, &classID)
			if err != nil {
				return nil, err
			}
		}

		item.Name = name.String
		item.StudentNo = studentNo.String
		item.Grade = grade.String
		item.EducationLevel = educationLevel.String
		item.SchoolSystem = schoolSystem.String
		item.IDCard = idCard.String
		item.AdmissionNo = admissionNo.String
		item.ExamNo = examNo.String
		if areaID.Valid {
			a := int(areaID.Int64)
			item.AreaID = &a
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			item.FacultyID = &f
		}
		if classID.Valid {
			c := int(classID.Int64)
			item.ClassID = &c
		}

		items = append(items, item)
	}

	return &model.OpenStudentResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
