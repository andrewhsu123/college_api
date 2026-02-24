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

// GetPersonInfo 获取人员完整信息（根据person_type返回不同结构）
func (r *PersonRepository) GetPersonInfo(personID int) (model.PersonInfo, error) {
	baseQuery := `
        SELECT id, customer_id, person_type, name, gender, mobile, email, avatar, status
        FROM persons
        WHERE id = ? AND deleted_at = 0
        LIMIT 1
    `

	var pID, universityID, personType, status int
	var name string
	var gender sql.NullInt64
	var mobile, email, avatar sql.NullString

	err := r.db.QueryRow(baseQuery, personID).Scan(
		&pID, &universityID, &personType, &name, &gender, &mobile, &email, &avatar, &status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	switch personType {
	case 1: // 学生
		info := &model.StudentPersonInfo{
			PersonID:     pID,
			PersonType:   personType,
			UniversityID: universityID,
			Name:         name,
			Status:       status,
		}
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
		r.fillStudentPersonInfo(info)
		return info, nil

	case 2, 3: // 政工、维修工
		info := &model.StaffPersonInfo{
			PersonID:     pID,
			PersonType:   personType,
			UniversityID: universityID,
			Name:         name,
			Status:       status,
		}
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
		r.fillStaffPersonInfo(info)
		return info, nil

	default:
		return nil, fmt.Errorf("unsupported person type: %d", personType)
	}
}

// fillStudentPersonInfo 填充学生扩展信息
func (r *PersonRepository) fillStudentPersonInfo(info *model.StudentPersonInfo) {
	query := `
        SELECT student_no, grade, area_id, education_level, school_system,
               id_card, admission_no, exam_no, enrollment_status, is_enrolled,
               college_id, faculty_id, profession_id, class_id
        FROM students
        WHERE person_id = ?
        LIMIT 1
    `

	var studentNo, grade, educationLevel, schoolSystem, idCard, admissionNo, examNo sql.NullString
	var areaID, enrollmentStatus, isEnrolled, collegeID, facultyID, professionID, classID sql.NullInt64

	err := r.db.QueryRow(query, info.PersonID).Scan(
		&studentNo, &grade, &areaID, &educationLevel, &schoolSystem,
		&idCard, &admissionNo, &examNo, &enrollmentStatus, &isEnrolled,
		&collegeID, &facultyID, &professionID, &classID,
	)
	if err != nil {
		return
	}

	if studentNo.Valid {
		info.StudentNo = studentNo.String
	}
	if grade.Valid {
		info.Grade = grade.String
	}
	if areaID.Valid {
		a := int(areaID.Int64)
		info.AreaID = &a
	}
	if educationLevel.Valid {
		info.EducationLevel = educationLevel.String
	}
	if schoolSystem.Valid {
		info.SchoolSystem = schoolSystem.String
	}
	if idCard.Valid {
		info.IDCard = idCard.String
	}
	if admissionNo.Valid {
		info.AdmissionNo = admissionNo.String
	}
	if examNo.Valid {
		info.ExamNo = examNo.String
	}
	if enrollmentStatus.Valid {
		e := int(enrollmentStatus.Int64)
		info.EnrollmentStatus = &e
	}
	if isEnrolled.Valid {
		i := int(isEnrolled.Int64)
		info.IsEnrolled = &i
	}
	if collegeID.Valid {
		c := int(collegeID.Int64)
		info.CollegeID = &c
	}
	if facultyID.Valid {
		f := int(facultyID.Int64)
		info.FacultyID = &f
	}
	if professionID.Valid {
		p := int(professionID.Int64)
		info.ProfessionID = &p
	}
	if classID.Valid {
		c := int(classID.Int64)
		info.ClassID = &c
	}
}

// fillStaffPersonInfo 填充政工/维修工扩展信息
func (r *PersonRepository) fillStaffPersonInfo(info *model.StaffPersonInfo) {
	query := `
        SELECT staff_no, department_id, college_id, faculty_id
        FROM staff
        WHERE person_id = ?
        LIMIT 1
    `

	var staffNo sql.NullString
	var departmentID, collegeID, facultyID sql.NullInt64

	err := r.db.QueryRow(query, info.PersonID).Scan(
		&staffNo, &departmentID, &collegeID, &facultyID,
	)
	if err != nil {
		return
	}

	if staffNo.Valid {
		info.StaffNo = staffNo.String
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
}

// GetStaffInfo 获取政工完整信息 (兼容旧接口)
func (r *PersonRepository) GetStaffInfo(personID int) (*model.StaffPersonInfo, error) {
	info, err := r.GetPersonInfo(personID)
	if err != nil {
		return nil, err
	}
	if staffInfo, ok := info.(*model.StaffPersonInfo); ok {
		return staffInfo, nil
	}
	return nil, fmt.Errorf("person is not staff type")
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
        SELECT id, username, email, mobile, avatar, customer_id, status
        FROM admin_users
        WHERE id = ? AND deleted_at = 0
        LIMIT 1
    `
	var info model.AdminUserInfo
	var email, mobile, avatar sql.NullString
	err := r.db.QueryRow(query, userID).Scan(
		&info.UserID, &info.Username, &email, &mobile, &avatar, &info.UniversityID, &info.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin user not found")
		}
		return nil, err
	}
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
	baseQuery := `
        SELECT p.id, p.customer_id, p.person_type, p.name, p.gender, 
               p.mobile, p.email, p.avatar, p.status
        FROM persons p
        WHERE p.customer_id = ? AND p.person_type = ? AND p.deleted_at = 0
    `

	args := []interface{}{req.UniversityID, req.PersonType}

	if isStaff {
		baseQuery += r.buildPermissionFilter(req.PersonType, len(visibleDeptIDs), len(managedPersonIDs))
		if len(managedPersonIDs) > 0 {
			for _, id := range managedPersonIDs {
				args = append(args, id)
			}
		}
		if len(visibleDeptIDs) > 0 {
			for i := 0; i < 4; i++ {
				for _, id := range visibleDeptIDs {
					args = append(args, id)
				}
			}
			if req.PersonType == 2 {
				for i := 0; i < 3; i++ {
					for _, id := range visibleDeptIDs {
						args = append(args, id)
					}
				}
			}
		}
	}

	baseQuery, args = r.addBasicFilters(baseQuery, args, req)
	baseQuery, args = r.addExtendFilters(baseQuery, args, req)

	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS t"
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count persons: %w", err)
	}

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

func (r *PersonRepository) buildPermissionFilter(personType int, deptCount, personCount int) string {
	if personCount == 0 && deptCount == 0 {
		return " AND 1=0"
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

		if personType == 1 {
			conditions = append(conditions, fmt.Sprintf(`EXISTS (
                SELECT 1 FROM students s
                WHERE s.person_id = p.id
                  AND (s.college_id IN (%s) OR s.faculty_id IN (%s) 
                       OR s.profession_id IN (%s) OR s.class_id IN (%s))
            )`, deptPlaceholders, deptPlaceholders, deptPlaceholders, deptPlaceholders))
		} else if personType == 2 {
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

func (r *PersonRepository) addExtendFilters(query string, args []interface{}, req *model.PersonListRequest) (string, []interface{}) {
	if req.PersonType == 1 {
		return r.addStudentFilters(query, args, req)
	} else if req.PersonType == 2 {
		return r.addStaffFilters(query, args, req)
	}
	return query, args
}

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
