package model

// OpenStaffRequest 开放接口政工查询请求
type OpenStaffRequest struct {
	UniversityID  int    `form:"university_id" binding:"required"`
	PersonIDs     string `form:"person_id"`      // 支持逗号分隔
	Names         string `form:"name"`           // 支持逗号分隔
	StaffNos      string `form:"staff_no"`       // 支持逗号分隔
	DepartmentID  *int   `form:"department_id"`  // 单个部门ID
	DepartmentIDs string `form:"department_ids"` // 支持逗号分隔多个部门ID
	CollegeID     *int   `form:"college_id"`     // 单个学院ID
	CollegeIDs    string `form:"college_ids"`    // 支持逗号分隔多个学院ID
	FacultyID     *int   `form:"faculty_id"`     // 单个系ID
	FacultyIDs    string `form:"faculty_ids"`    // 支持逗号分隔多个系ID
	WithContact   *int   `form:"with_contact"`   // 1=显示手机、邮箱、性别、状态
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
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
	CollegeID        *int   `form:"college_id"`     // 单个学院ID
	CollegeIDs       string `form:"college_ids"`    // 支持逗号分隔多个学院ID
	FacultyID        *int   `form:"faculty_id"`     // 单个系ID
	FacultyIDs       string `form:"faculty_ids"`    // 支持逗号分隔多个系ID
	ProfessionID     *int   `form:"profession_id"`  // 单个专业ID
	ProfessionIDs    string `form:"profession_ids"` // 支持逗号分隔多个专业ID
	ClassID          *int   `form:"class_id"`       // 单个班级ID
	ClassIDs         string `form:"class_ids"`      // 支持逗号分隔多个班级ID
	WithContact      *int   `form:"with_contact"`   // 1=显示手机、邮箱、性别、状态
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

// OpenRolePersonsRequest 根据角色查询人员请求
type OpenRolePersonsRequest struct {
	UniversityID int    `form:"university_id" binding:"required"`
	RoleIDs      string `form:"role_ids"`  // 支持逗号分隔，如 68,63
	RoleName     string `form:"role_name"` // 按角色名称查询（与 role_ids 二选一）
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// OpenRolePersonItem 角色人员详情
type OpenRolePersonItem struct {
	PersonID       int     `json:"person_id"`
	PersonType     int     `json:"person_type"`
	Name           string  `json:"name"`
	Mobile         *string `json:"mobile"`
	Status         int     `json:"status"`
	StudentNo      *string `json:"student_no,omitempty"`
	StaffNo        *string `json:"staff_no,omitempty"`
	CollegeID      *int    `json:"college_id,omitempty"`
	CollegeName    *string `json:"college_name,omitempty"`
	FacultyID      *int    `json:"faculty_id,omitempty"`
	FacultyName    *string `json:"faculty_name,omitempty"`
	DepartmentID   *int    `json:"department_id,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"`
	ProfessionID   *int    `json:"profession_id,omitempty"`
	ProfessionName *string `json:"profession_name,omitempty"`
	ClassID        *int    `json:"class_id,omitempty"`
	ClassName      *string `json:"class_name,omitempty"`
}

// OpenRolePersonsResponse 根据角色查询人员响应
type OpenRolePersonsResponse struct {
	Items    []OpenRolePersonItem `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// OpenCollegeRequest 开放接口学校查询请求
type OpenCollegeRequest struct {
	Username string `form:"username"`
	Mobile   string `form:"mobile"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// OpenCollegeItem 学校查询结果项
type OpenCollegeItem struct {
	CustomerID int    `json:"university_id"`
	Username   string `json:"username"`
	Mobile     string `json:"mobile"`
	Status     int    `json:"status"`
}

// OpenCollegeResponse 学校查询响应
type OpenCollegeResponse struct {
	Items    []OpenCollegeItem `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// OpenCampusAreaRequest 开放接口校区查询请求
type OpenCampusAreaRequest struct {
	UniversityID int `form:"university_id" binding:"required"`
}

// OpenCampusAreaItem 校区查询结果项
type OpenCampusAreaItem struct {
	ID       int    `json:"id"`
	AreaName string `json:"area_name"`
}

// OpenDepartmentListRequest 开放接口部门列表查询请求
type OpenDepartmentListRequest struct {
	UniversityID int `form:"university_id" binding:"required"`
}

// OpenDepartmentListItem 部门列表查询结果项
type OpenDepartmentListItem struct {
	ID             int    `json:"id"`
	ParentID       int    `json:"parent_id"`
	DepartmentName string `json:"department_name"`
	DepartmentType int    `json:"department_type"`
}

// OpenStaffByOrgRequest 按组织机构查询政工请求（OR条件）
type OpenStaffByOrgRequest struct {
	UniversityID  int    `form:"university_id" binding:"required"` // 学校ID（必填）
	DepartmentIDs string `form:"department_ids"`                   // 部门ID列表，逗号分隔
	CollegeIDs    string `form:"college_ids"`                      // 学院ID列表，逗号分隔
	FacultyIDs    string `form:"faculty_ids"`                      // 系ID列表，逗号分隔
	WithContact   *int   `form:"with_contact"`                     // 1=显示手机、邮箱、性别、状态
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
}

// OpenStudentByOrgRequest 按组织机构查询学生请求（OR条件）
type OpenStudentByOrgRequest struct {
	UniversityID  int    `form:"university_id" binding:"required"` // 学校ID（必填）
	CollegeIDs    string `form:"college_ids"`                      // 学院ID列表，逗号分隔
	FacultyIDs    string `form:"faculty_ids"`                      // 系ID列表，逗号分隔
	ProfessionIDs string `form:"profession_ids"`                   // 专业ID列表，逗号分隔
	ClassIDs      string `form:"class_ids"`                        // 班级ID列表，逗号分隔
	WithContact   *int   `form:"with_contact"`                     // 1=显示手机、邮箱、性别、状态
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
}
