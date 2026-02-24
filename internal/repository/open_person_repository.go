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

	if req.DepartmentID != nil {
		query += " AND s.department_id = ?"
		args = append(args, *req.DepartmentID)
	}
	if req.CollegeID != nil {
		query += " AND s.college_id = ?"
		args = append(args, *req.CollegeID)
	}
	if req.FacultyID != nil {
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
	if req.CollegeID != nil {
		query += " AND s.college_id = ?"
		args = append(args, *req.CollegeID)
	}
	if req.FacultyID != nil {
		query += " AND s.faculty_id = ?"
		args = append(args, *req.FacultyID)
	}
	if req.ProfessionID != nil {
		query += " AND s.profession_id = ?"
		args = append(args, *req.ProfessionID)
	}
	if req.ClassID != nil {
		query += " AND s.class_id = ?"
		args = append(args, *req.ClassID)
	}

	return query, args
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
