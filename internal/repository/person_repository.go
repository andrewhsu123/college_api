package repository

import (
	"college_api/internal/model"
	"database/sql"
	"fmt"
)

// PersonRepository 人员数据访问层
type PersonRepository struct {
	db *sql.DB
}

// NewPersonRepository 创建人员仓库
func NewPersonRepository(db *sql.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

// GetStaffInfo 获取政工完整信息
func (r *PersonRepository) GetStaffInfo(personID int) (*model.StaffInfo, error) {
	query := `
        SELECT 
            p.id as person_id,
            p.person_type,
            p.name,
            p.gender,
            p.mobile,
            p.email,
            p.avatar,
            p.status,
            s.staff_no,
            s.university_id,
            s.department_id,
            s.college_id,
            s.faculty_id
        FROM persons p
        LEFT JOIN staff s ON p.id = s.person_id
        WHERE p.id = ? AND p.deleted_at = 0
        LIMIT 1
    `

	var info model.StaffInfo
	var gender sql.NullInt64
	var mobile, email, avatar, staffNo sql.NullString
	var universityID sql.NullInt64
	var departmentID, collegeID, facultyID sql.NullInt64

	err := r.db.QueryRow(query, personID).Scan(
		&info.PersonID,
		&info.PersonType,
		&info.Name,
		&gender,
		&mobile,
		&email,
		&avatar,
		&info.Status,
		&staffNo,
		&universityID,
		&departmentID,
		&collegeID,
		&facultyID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	// 处理可空字段
	if gender.Valid {
		g := int(gender.Int64)
		info.Gender = &g
	}
	if mobile.Valid {
		info.Mobile = mobile.String
	}
	if email.Valid {
		info.Email = email.String
	}
	if avatar.Valid {
		info.Avatar = avatar.String
	}
	if staffNo.Valid {
		info.StaffNo = staffNo.String
	}
	if universityID.Valid {
		info.UniversityID = int(universityID.Int64)
	}
	if departmentID.Valid {
		d := int(departmentID.Int64)
		info.DepartmentID = &d
	}
	if collegeID.Valid {
		c := int(collegeID.Int64)
		info.CollegeID = &c
	}
	if facultyID.Valid {
		f := int(facultyID.Int64)
		info.FacultyID = &f
	}

	return &info, nil
}

// GetDepartmentName 获取部门名称
func (r *PersonRepository) GetDepartmentName(deptID int) (string, error) {
	query := `SELECT department_name FROM departments WHERE id = ? AND deleted_at = 0 LIMIT 1`

	var name string
	err := r.db.QueryRow(query, deptID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("department not found")
		}
		return "", err
	}

	return name, nil
}

// GetAdminUserInfo 获取学校管理员信息
func (r *PersonRepository) GetAdminUserInfo(userID int) (*model.AdminUserInfo, error) {
	query := `
        SELECT 
            id,
            username,
            email,
            mobile,
            avatar,
            customer_id,
            status
        FROM admin_users
        WHERE id = ? AND deleted_at = 0
        LIMIT 1
    `

	var info model.AdminUserInfo
	var email, mobile, avatar sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&info.UserID,
		&info.Username,
		&email,
		&mobile,
		&avatar,
		&info.UniversityID,
		&info.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin user not found")
		}
		return nil, err
	}

	// 处理可空字段
	if email.Valid {
		info.Email = email.String
	}
	if mobile.Valid {
		info.Mobile = mobile.String
	}
	if avatar.Valid {
		info.Avatar = avatar.String
	}

	return &info, nil
}

// GetPersonList 查询人员列表（基础信息）
func (r *PersonRepository) GetPersonList(req *model.PersonListRequest, visibleDeptIDs, managedPersonIDs []int, isStaff bool) ([]model.PersonWithExtend, int64, error) {
	// 构建基础查询
	baseQuery := `
        SELECT p.id, p.customer_id, p.person_type, p.name, p.gender, 
               p.mobile, p.email, p.avatar, p.status
        FROM persons p
        WHERE p.customer_id = ? AND p.person_type = ? AND p.deleted_at = 0
    `

	args := []interface{}{req.UniversityID, req.PersonType}

	// 添加权限过滤（政工人员）
	if isStaff {
		baseQuery += r.buildPermissionFilter(req.PersonType, len(visibleDeptIDs), len(managedPersonIDs))
		if len(managedPersonIDs) > 0 {
			for _, id := range managedPersonIDs {
				args = append(args, id)
			}
		}
		if len(visibleDeptIDs) > 0 {
			for i := 0; i < 4; i++ { // 学生有4个部门字段
				for _, id := range visibleDeptIDs {
					args = append(args, id)
				}
			}
			if req.PersonType == 2 { // 政工有3个部门字段
				for i := 0; i < 3; i++ {
					for _, id := range visibleDeptIDs {
						args = append(args, id)
					}
				}
			}
		}
	}

	// 添加基础字段搜索条件
	baseQuery, args = r.addBasicFilters(baseQuery, args, req)

	// 添加扩展字段搜索条件
	baseQuery, args = r.addExtendFilters(baseQuery, args, req)

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS t"
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count persons: %w", err)
	}

	// 分页查询
	baseQuery += " ORDER BY p.id DESC LIMIT ? OFFSET ?"
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query persons: %w", err)
	}
	defer rows.Close()

	var persons []model.PersonWithExtend
	for rows.Next() {
		var p model.PersonWithExtend
		var gender sql.NullInt64
		var mobile, email, avatar sql.NullString

		err := rows.Scan(&p.ID, &p.CustomerID, &p.PersonType, &p.Name, &gender,
			&mobile, &email, &avatar, &p.Status)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan person: %w", err)
		}

		if gender.Valid {
			g := int(gender.Int64)
			p.Gender = &g
		}
		if mobile.Valid {
			p.Mobile = mobile.String
		}
		if email.Valid {
			p.Email = email.String
		}
		if avatar.Valid {
			p.Avatar = avatar.String
		}

		persons = append(persons, p)
	}

	return persons, total, nil
}

// buildPermissionFilter 构建权限过滤条件
func (r *PersonRepository) buildPermissionFilter(personType int, deptCount, personCount int) string {
	if personCount == 0 && deptCount == 0 {
		return " AND 1=0" // 无权限
	}

	filter := " AND ("
	conditions := []string{}

	if personCount > 0 {
		placeholders := ""
		for i := 0; i < personCount; i++ {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
		}
		conditions = append(conditions, "p.id IN ("+placeholders+")")
	}

	if deptCount > 0 {
		deptPlaceholders := ""
		for i := 0; i < deptCount; i++ {
			if i > 0 {
				deptPlaceholders += ","
			}
			deptPlaceholders += "?"
		}

		if personType == 1 { // 学生
			conditions = append(conditions, fmt.Sprintf(`EXISTS (
                SELECT 1 FROM students s
                WHERE s.person_id = p.id
                  AND (s.college_id IN (%s) OR s.faculty_id IN (%s) 
                       OR s.profession_id IN (%s) OR s.class_id IN (%s))
            )`, deptPlaceholders, deptPlaceholders, deptPlaceholders, deptPlaceholders))
		} else if personType == 2 { // 政工
			conditions = append(conditions, fmt.Sprintf(`EXISTS (
                SELECT 1 FROM staff st
                WHERE st.person_id = p.id
                  AND (st.department_id IN (%s) OR st.college_id IN (%s) OR st.faculty_id IN (%s))
            )`, deptPlaceholders, deptPlaceholders, deptPlaceholders))
		}
	}

	for i, cond := range conditions {
		if i > 0 {
			filter += " OR "
		}
		filter += cond
	}
	filter += ")"

	return filter
}

// addBasicFilters 添加基础字段过滤条件
func (r *PersonRepository) addBasicFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	if req.Name != "" {
		query += " AND p.name LIKE ?"
		args = append(args, "%"+req.Name+"%")
	}
	if req.Mobile != "" {
		query += " AND p.mobile = ?"
		args = append(args, req.Mobile)
	}
	if req.Email != "" {
		query += " AND p.email = ?"
		args = append(args, req.Email)
	}
	if req.Gender != nil {
		query += " AND p.gender = ?"
		args = append(args, *req.Gender)
	}
	if req.Status != nil {
		query += " AND p.status = ?"
		args = append(args, *req.Status)
	}
	return query, args
}

// addExtendFilters 添加扩展字段过滤条件
func (r *PersonRepository) addExtendFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	if req.PersonType == 1 {
		return r.addStudentFilters(query, args, req)
	} else if req.PersonType == 2 {
		return r.addStaffFilters(query, args, req)
	}
	return query, args
}

// addStudentFilters 添加学生扩展字段过滤
func (r *PersonRepository) addStudentFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	conditions := []string{}
	studentArgs := []interface{}{}

	if req.StudentNo != "" {
		conditions = append(conditions, "s.student_no LIKE ?")
		studentArgs = append(studentArgs, "%"+req.StudentNo+"%")
	}
	if req.AreaID != nil {
		conditions = append(conditions, "s.area_id = ?")
		studentArgs = append(studentArgs, *req.AreaID)
	}
	if req.Grade != "" {
		conditions = append(conditions, "s.grade = ?")
		studentArgs = append(studentArgs, req.Grade)
	}
	if req.EducationLevel != "" {
		conditions = append(conditions, "s.education_level = ?")
		studentArgs = append(studentArgs, req.EducationLevel)
	}
	if req.SchoolSystem != "" {
		conditions = append(conditions, "s.school_system = ?")
		studentArgs = append(studentArgs, req.SchoolSystem)
	}
	if req.IDCard != "" {
		conditions = append(conditions, "s.id_card = ?")
		studentArgs = append(studentArgs, req.IDCard)
	}
	if req.AdmissionNo != "" {
		conditions = append(conditions, "s.admission_no = ?")
		studentArgs = append(studentArgs, req.AdmissionNo)
	}
	if req.ExamNo != "" {
		conditions = append(conditions, "s.exam_no = ?")
		studentArgs = append(studentArgs, req.ExamNo)
	}
	if req.EnrollmentStatus != nil {
		conditions = append(conditions, "s.enrollment_status = ?")
		studentArgs = append(studentArgs, *req.EnrollmentStatus)
	}
	if req.IsEnrolled != nil {
		conditions = append(conditions, "s.is_enrolled = ?")
		studentArgs = append(studentArgs, *req.IsEnrolled)
	}
	if req.CollegeID != nil {
		conditions = append(conditions, "s.college_id = ?")
		studentArgs = append(studentArgs, *req.CollegeID)
	}
	if req.FacultyID != nil {
		conditions = append(conditions, "s.faculty_id = ?")
		studentArgs = append(studentArgs, *req.FacultyID)
	}
	if req.ProfessionID != nil {
		conditions = append(conditions, "s.profession_id = ?")
		studentArgs = append(studentArgs, *req.ProfessionID)
	}
	if req.ClassID != nil {
		conditions = append(conditions, "s.class_id = ?")
		studentArgs = append(studentArgs, *req.ClassID)
	}

	if len(conditions) > 0 {
		filter := " AND EXISTS (SELECT 1 FROM students s WHERE s.person_id = p.id"
		for _, cond := range conditions {
			filter += " AND " + cond
		}
		filter += ")"
		query += filter
		args = append(args, studentArgs...)
	}

	return query, args
}

// addStaffFilters 添加政工扩展字段过滤
func (r *PersonRepository) addStaffFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	conditions := []string{}
	staffArgs := []interface{}{}

	if req.StaffNo != "" {
		conditions = append(conditions, "st.staff_no LIKE ?")
		staffArgs = append(staffArgs, "%"+req.StaffNo+"%")
	}
	if req.DepartmentID != nil {
		conditions = append(conditions, "st.department_id = ?")
		staffArgs = append(staffArgs, *req.DepartmentID)
	}
	if req.CollegeID != nil {
		conditions = append(conditions, "st.college_id = ?")
		staffArgs = append(staffArgs, *req.CollegeID)
	}
	if req.FacultyID != nil {
		conditions = append(conditions, "st.faculty_id = ?")
		staffArgs = append(staffArgs, *req.FacultyID)
	}

	if len(conditions) > 0 {
		filter := " AND EXISTS (SELECT 1 FROM staff st WHERE st.person_id = p.id"
		for _, cond := range conditions {
			filter += " AND " + cond
		}
		filter += ")"
		query += filter
		args = append(args, staffArgs...)
	}

	return query, args
}

// GetStudentExtendInfo 批量查询学生扩展信息
func (r *PersonRepository) GetStudentExtendInfo(personIDs []int, req *model.PersonListRequest) (map[int]*model.StudentExtend, error) {
	if len(personIDs) == 0 {
		return make(map[int]*model.StudentExtend), nil
	}

	query := `
        SELECT person_id, area_id, student_no, grade, education_level,
               school_system, id_card, admission_no, exam_no,
               enrollment_status, is_enrolled, college_id, faculty_id,
               profession_id, class_id
        FROM students
        WHERE person_id IN (`

	args := []interface{}{}
	for i, id := range personIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	// 添加过滤条件
	query, args = r.addStudentExtendFilters(query, args, req)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query student extend info: %w", err)
	}
	defer rows.Close()

	extendMap := make(map[int]*model.StudentExtend)
	for rows.Next() {
		var ext model.StudentExtend
		var areaID, facultyID, classID sql.NullInt64

		err := rows.Scan(&ext.PersonID, &areaID, &ext.StudentNo, &ext.Grade, &ext.EducationLevel,
			&ext.SchoolSystem, &ext.IDCard, &ext.AdmissionNo, &ext.ExamNo,
			&ext.EnrollmentStatus, &ext.IsEnrolled, &ext.CollegeID, &facultyID,
			&ext.ProfessionID, &classID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student extend: %w", err)
		}

		if areaID.Valid {
			a := int(areaID.Int64)
			ext.AreaID = &a
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			ext.FacultyID = &f
		}
		if classID.Valid {
			c := int(classID.Int64)
			ext.ClassID = &c
		}

		extendMap[ext.PersonID] = &ext
	}

	return extendMap, nil
}

// addStudentExtendFilters 添加学生扩展信息查询的过滤条件
func (r *PersonRepository) addStudentExtendFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	if req.StudentNo != "" {
		query += " AND student_no LIKE ?"
		args = append(args, "%"+req.StudentNo+"%")
	}
	if req.AreaID != nil {
		query += " AND area_id = ?"
		args = append(args, *req.AreaID)
	}
	if req.Grade != "" {
		query += " AND grade = ?"
		args = append(args, req.Grade)
	}
	if req.EducationLevel != "" {
		query += " AND education_level = ?"
		args = append(args, req.EducationLevel)
	}
	if req.SchoolSystem != "" {
		query += " AND school_system = ?"
		args = append(args, req.SchoolSystem)
	}
	if req.IDCard != "" {
		query += " AND id_card = ?"
		args = append(args, req.IDCard)
	}
	if req.AdmissionNo != "" {
		query += " AND admission_no = ?"
		args = append(args, req.AdmissionNo)
	}
	if req.ExamNo != "" {
		query += " AND exam_no = ?"
		args = append(args, req.ExamNo)
	}
	if req.EnrollmentStatus != nil {
		query += " AND enrollment_status = ?"
		args = append(args, *req.EnrollmentStatus)
	}
	if req.IsEnrolled != nil {
		query += " AND is_enrolled = ?"
		args = append(args, *req.IsEnrolled)
	}
	if req.CollegeID != nil {
		query += " AND college_id = ?"
		args = append(args, *req.CollegeID)
	}
	if req.FacultyID != nil {
		query += " AND faculty_id = ?"
		args = append(args, *req.FacultyID)
	}
	if req.ProfessionID != nil {
		query += " AND profession_id = ?"
		args = append(args, *req.ProfessionID)
	}
	if req.ClassID != nil {
		query += " AND class_id = ?"
		args = append(args, *req.ClassID)
	}
	return query, args
}

// GetStaffExtendInfo 批量查询政工扩展信息
func (r *PersonRepository) GetStaffExtendInfo(personIDs []int, req *model.PersonListRequest) (map[int]*model.StaffExtend, error) {
	if len(personIDs) == 0 {
		return make(map[int]*model.StaffExtend), nil
	}

	query := `
        SELECT person_id, staff_no, department_id, college_id, faculty_id
        FROM staff
        WHERE person_id IN (`

	args := []interface{}{}
	for i, id := range personIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	// 添加过滤条件
	query, args = r.addStaffExtendFilters(query, args, req)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query staff extend info: %w", err)
	}
	defer rows.Close()

	extendMap := make(map[int]*model.StaffExtend)
	for rows.Next() {
		var ext model.StaffExtend
		var departmentID, collegeID, facultyID sql.NullInt64

		err := rows.Scan(&ext.PersonID, &ext.StaffNo, &departmentID, &collegeID, &facultyID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan staff extend: %w", err)
		}

		if departmentID.Valid {
			d := int(departmentID.Int64)
			ext.DepartmentID = &d
		}
		if collegeID.Valid {
			c := int(collegeID.Int64)
			ext.CollegeID = &c
		}
		if facultyID.Valid {
			f := int(facultyID.Int64)
			ext.FacultyID = &f
		}

		extendMap[ext.PersonID] = &ext
	}

	return extendMap, nil
}

// addStaffExtendFilters 添加政工扩展信息查询的过滤条件
func (r *PersonRepository) addStaffExtendFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	if req.StaffNo != "" {
		query += " AND staff_no LIKE ?"
		args = append(args, "%"+req.StaffNo+"%")
	}
	if req.DepartmentID != nil {
		query += " AND department_id = ?"
		args = append(args, *req.DepartmentID)
	}
	if req.CollegeID != nil {
		query += " AND college_id = ?"
		args = append(args, *req.CollegeID)
	}
	if req.FacultyID != nil {
		query += " AND faculty_id = ?"
		args = append(args, *req.FacultyID)
	}
	return query, args
}
