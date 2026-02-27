package model

// OpenStaffRequest 开放接口政工查询请求
type OpenStaffRequest struct {
	UniversityID int    `form:"university_id" binding:"required"`
	PersonIDs    string `form:"person_id"` // 支持逗号分隔
	Names        string `form:"name"`      // 支持逗号分隔
	StaffNos     string `form:"staff_no"`  // 支持逗号分隔
	DepartmentID *int   `form:"department_id"`
	CollegeID    *int   `form:"college_id"`
	FacultyID    *int   `form:"faculty_id"`
	WithContact  *int   `form:"with_contact"` // 1=显示手机、邮箱、性别、状态
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// OpenStaffItem 政工查询结果项
type OpenStaffItem struct {
	PersonID     int     `json:"person_id"`
	Name         string  `json:"name"`
	StaffNo      string  `json:"staff_no"`
	DepartmentID *int    `json:"department_id"`
	CollegeID    *int    `json:"college_id"`
	FacultyID    *int    `json:"faculty_id"`
	Mobile       *string `json:"mobile,omitempty"`
	Email        *string `json:"email,omitempty"`
	Gender       *int    `json:"gender,omitempty"`
	Status       *int    `json:"status,omitempty"`
}

// OpenStaffResponse 政工查询响应
type OpenStaffResponse struct {
	Items    []OpenStaffItem `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// OpenStudentRequest 开放接口学生查询请求
type OpenStudentRequest struct {
	UniversityID     int    `form:"university_id" binding:"required"`
	PersonIDs        string `form:"person_id"`  // 支持逗号分隔
	Names            string `form:"name"`       // 支持逗号分隔
	StudentNos       string `form:"student_no"` // 支持逗号分隔
	AreaID           *int   `form:"area_id"`
	Grade            string `form:"grade"`
	EducationLevel   string `form:"education_level"`
	SchoolSystem     string `form:"school_system"`
	IDCard           string `form:"id_card"`
	AdmissionNo      string `form:"admission_no"`
	ExamNo           string `form:"exam_no"`
	EnrollmentStatus *int   `form:"enrollment_status"`
	IsEnrolled       *int   `form:"is_enrolled"`
	CollegeID        *int   `form:"college_id"`
	FacultyID        *int   `form:"faculty_id"`
	ProfessionID     *int   `form:"profession_id"`
	ClassID          *int   `form:"class_id"`
	WithContact      *int   `form:"with_contact"` // 1=显示手机、邮箱、性别、状态
	Page             int    `form:"page"`
	PageSize         int    `form:"page_size"`
}

// OpenStudentItem 学生查询结果项
type OpenStudentItem struct {
	PersonID         int     `json:"person_id"`
	Name             string  `json:"name"`
	StudentNo        string  `json:"student_no"`
	AreaID           *int    `json:"area_id"`
	Grade            string  `json:"grade"`
	EducationLevel   string  `json:"education_level"`
	SchoolSystem     string  `json:"school_system"`
	IDCard           string  `json:"id_card"`
	AdmissionNo      string  `json:"admission_no"`
	ExamNo           string  `json:"exam_no"`
	EnrollmentStatus int     `json:"enrollment_status"`
	IsEnrolled       int     `json:"is_enrolled"`
	CollegeID        int     `json:"college_id"`
	FacultyID        *int    `json:"faculty_id"`
	ProfessionID     int     `json:"profession_id"`
	ClassID          *int    `json:"class_id"`
	Mobile           *string `json:"mobile,omitempty"`
	Email            *string `json:"email,omitempty"`
	Gender           *int    `json:"gender,omitempty"`
	Status           *int    `json:"status,omitempty"`
}

// OpenStudentResponse 学生查询响应
type OpenStudentResponse struct {
	Items    []OpenStudentItem `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// OpenManagePersonsRequest 查询管辖某人员的管理者请求
type OpenManagePersonsRequest struct {
	UniversityID int `form:"university_id" binding:"required"`
	PersonID     int `form:"person_id" binding:"required"`
}

// OpenManagePersonItem 管辖人员的管理者信息
type OpenManagePersonItem struct {
	PersonID   int                    `json:"person_id"`
	PersonName string                 `json:"person_name"`
	Roles      []OpenManagePersonRole `json:"roles"`
}

// OpenManagePersonRole 管理者的角色信息
type OpenManagePersonRole struct {
	RoleID      int                          `json:"role_id"`
	RoleName    string                       `json:"role_name"`
	Departments []OpenManagePersonDepartment `json:"departments"`
}

// OpenManagePersonDepartment 角色管辖的机构信息
type OpenManagePersonDepartment struct {
	DepartmentID   int    `json:"department_id"`
	DepartmentName string `json:"department_name"`
}

// OpenManagePersonsResponse 查询管辖某人员的管理者响应
type OpenManagePersonsResponse struct {
	ManagePersons []OpenManagePersonItem `json:"manage_persons"`
}

// OpenRoleRequest 开放接口角色查询请求
type OpenRoleRequest struct {
	UniversityID int `form:"university_id" binding:"required"`
}

// OpenRoleItem 角色查询结果项
type OpenRoleItem struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id"`
	Name        string `json:"name"`
	Permissions string `json:"permissions"`
}
