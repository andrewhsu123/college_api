package model

// Person 人员基础模型
type Person struct {
	ID         int    `json:"id" db:"id"`
	CustomerID int    `json:"customer_id" db:"customer_id"` // 学校编号
	PersonType int    `json:"person_type" db:"person_type"` // 1=学生 2=政工 3=维修工
	Name       string `json:"name" db:"name"`
	Gender     *int   `json:"gender" db:"gender"` // 1=男 2=女
	Mobile     string `json:"mobile" db:"mobile"`
	Email      string `json:"email" db:"email"`
	Avatar     string `json:"avatar" db:"avatar"`
	Status     int    `json:"status" db:"status"` // 1=正常 2=禁用
}

// AdminUser 学校管理员模型
type AdminUser struct {
	ID           int    `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	Email        string `json:"email" db:"email"`
	Mobile       string `json:"mobile" db:"mobile"`
	Avatar       string `json:"avatar" db:"avatar"`
	CustomerID   int    `json:"customer_id" db:"customer_id"`
	DepartmentID int    `json:"department_id" db:"department_id"`
	Status       int    `json:"status" db:"status"` // 1=正常 2=禁用
}

// AdminUserInfo 学校管理员完整信息
type AdminUserInfo struct {
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Mobile         string `json:"mobile"`
	Avatar         string `json:"avatar"`
	UniversityID   int    `json:"university_id"`   // customer_id 改名为 university_id
	UniversityName string `json:"university_name"` // 大学名称
	Status         int    `json:"status"`
}

// Staff 政工扩展信息
type Staff struct {
	ID           int    `json:"id" db:"id"`
	UniversityID int    `json:"university_id" db:"university_id"`
	PersonID     int    `json:"person_id" db:"person_id"`
	StaffNo      string `json:"staff_no" db:"staff_no"`
	Name         string `json:"name" db:"name"`
	DepartmentID *int   `json:"department_id" db:"department_id"`
	CollegeID    *int   `json:"college_id" db:"college_id"`
	FacultyID    *int   `json:"faculty_id" db:"faculty_id"`
}

// PersonInfo 人员完整信息接口（用于统一处理）
type PersonInfo interface {
	GetPersonID() int
	GetPersonType() int
	GetUniversityID() int
	SetUniversityName(name string)
	SetManagedRoles(roles []ManagedRole)
	SetManagedMenu(menu []int)
	SetSelfDepartment(deptIDs []int)
	SetSelfRoles(roleIDs []int)
}

// StaffPersonInfo 政工/维修工信息（person_type=2,3）
type StaffPersonInfo struct {
	PersonID       int           `json:"person_id"`
	PersonType     int           `json:"person_type"`
	UniversityID   int           `json:"university_id"`
	UniversityName string        `json:"university_name"`
	Name           string        `json:"name"`
	Gender         *int          `json:"gender"`
	Mobile         string        `json:"mobile"`
	Email          string        `json:"email"`
	Avatar         string        `json:"avatar"`
	Status         int           `json:"status"`
	StaffNo        string        `json:"staff_no"`
	DepartmentID   *int          `json:"department_id"`
	DepartmentName *string       `json:"department_name"`
	CollegeID      *int          `json:"college_id"`
	CollegeName    *string       `json:"college_name"`
	FacultyID      *int          `json:"faculty_id"`
	FacultyName    *string       `json:"faculty_name"`
	ManagedRoles   []ManagedRole `json:"managed_roles"`
	ManagedMenu    []int         `json:"managed_menu"`
	SelfDepartment []int         `json:"self_department"` // 我的机构列表
	SelfRoles      []int         `json:"self_roles"`      // 我的角色列表
}

func (s *StaffPersonInfo) GetPersonID() int                    { return s.PersonID }
func (s *StaffPersonInfo) GetPersonType() int                  { return s.PersonType }
func (s *StaffPersonInfo) GetUniversityID() int                { return s.UniversityID }
func (s *StaffPersonInfo) SetUniversityName(name string)       { s.UniversityName = name }
func (s *StaffPersonInfo) SetManagedRoles(roles []ManagedRole) { s.ManagedRoles = roles }
func (s *StaffPersonInfo) SetManagedMenu(menu []int)           { s.ManagedMenu = menu }
func (s *StaffPersonInfo) SetSelfDepartment(deptIDs []int)     { s.SelfDepartment = deptIDs }
func (s *StaffPersonInfo) SetSelfRoles(roleIDs []int)          { s.SelfRoles = roleIDs }

// StudentPersonInfo 学生信息（person_type=1）
type StudentPersonInfo struct {
	PersonID         int           `json:"person_id"`
	PersonType       int           `json:"person_type"`
	UniversityID     int           `json:"university_id"`
	UniversityName   string        `json:"university_name"`
	Name             string        `json:"name"`
	Gender           *int          `json:"gender"`
	Mobile           string        `json:"mobile"`
	Email            string        `json:"email"`
	Avatar           string        `json:"avatar"`
	Status           int           `json:"status"`
	StudentNo        string        `json:"student_no"`
	Grade            string        `json:"grade"`
	AreaID           *int          `json:"area_id"`
	EducationLevel   string        `json:"education_level"`
	SchoolSystem     string        `json:"school_system"`
	IDCard           string        `json:"id_card"`
	AdmissionNo      string        `json:"admission_no"`
	ExamNo           string        `json:"exam_no"`
	EnrollmentStatus *int          `json:"enrollment_status"`
	IsEnrolled       *int          `json:"is_enrolled"`
	CollegeID        *int          `json:"college_id"`
	CollegeName      *string       `json:"college_name"`
	FacultyID        *int          `json:"faculty_id"`
	FacultyName      *string       `json:"faculty_name"`
	ProfessionID     *int          `json:"profession_id"`
	ProfessionName   *string       `json:"profession_name"`
	ClassID          *int          `json:"class_id"`
	ClassName        *string       `json:"class_name"`
	ManagedRoles     []ManagedRole `json:"managed_roles"`
	ManagedMenu      []int         `json:"managed_menu"`
	SelfDepartment   []int         `json:"self_department"` // 我的机构列表
	SelfRoles        []int         `json:"self_roles"`      // 我的角色列表
}

func (s *StudentPersonInfo) GetPersonID() int                    { return s.PersonID }
func (s *StudentPersonInfo) GetPersonType() int                  { return s.PersonType }
func (s *StudentPersonInfo) GetUniversityID() int                { return s.UniversityID }
func (s *StudentPersonInfo) SetUniversityName(name string)       { s.UniversityName = name }
func (s *StudentPersonInfo) SetManagedRoles(roles []ManagedRole) { s.ManagedRoles = roles }
func (s *StudentPersonInfo) SetManagedMenu(menu []int)           { s.ManagedMenu = menu }
func (s *StudentPersonInfo) SetSelfDepartment(deptIDs []int)     { s.SelfDepartment = deptIDs }
func (s *StudentPersonInfo) SetSelfRoles(roleIDs []int)          { s.SelfRoles = roleIDs }

// ManagedRole 管辖角色信息
type ManagedRole struct {
	ID          int                 `json:"id"`
	ParentID    int                 `json:"parent_id"`
	ParentName  string              `json:"parent_name"`
	Name        string              `json:"name"`
	Departments []ManagedDepartment `json:"departments"`
}

// ManagedDepartment 管辖机构信息
type ManagedDepartment struct {
	ID             int    `json:"id"`
	ParentID       int    `json:"parent_id"`
	DepartmentName string `json:"department_name"`
	DepartmentType int    `json:"department_type"`
	Status         int    `json:"status"`
}

// StaffInfo 政工完整信息 - 保留兼容
type StaffInfo = StaffPersonInfo

// PersonListRequest 人员列表查询请求
type PersonListRequest struct {
	UniversityID int  `form:"university_id" binding:"required"`
	PersonType   int  `form:"person_type" binding:"required"`
	Page         int  `form:"page"`
	PageSize     int  `form:"page_size"`
	WithExtend   bool `form:"with_extend"`

	// persons表字段搜索
	Name   string `form:"name"`
	Mobile string `form:"mobile"`
	Email  string `form:"email"`
	Gender *int   `form:"gender"`
	Status *int   `form:"status"`

	// 政工扩展字段搜索
	StaffNo      string `form:"staff_no"`
	DepartmentID *int   `form:"department_id"`

	// 学生扩展字段搜索
	StudentNo        string `form:"student_no"`
	AreaID           *int   `form:"area_id"`
	Grade            string `form:"grade"`
	EducationLevel   string `form:"education_level"`
	SchoolSystem     string `form:"school_system"`
	IDCard           string `form:"id_card"`
	AdmissionNo      string `form:"admission_no"`
	ExamNo           string `form:"exam_no"`
	EnrollmentStatus *int   `form:"enrollment_status"`
	IsEnrolled       *int   `form:"is_enrolled"`

	// 共用的组织字段
	CollegeID    *int `form:"college_id"`
	FacultyID    *int `form:"faculty_id"`
	ProfessionID *int `form:"profession_id"`
	ClassID      *int `form:"class_id"`
}

// PersonWithExtend 人员信息（含扩展）
type PersonWithExtend struct {
	ID         int    `json:"id"`
	CustomerID int    `json:"customer_id"`
	PersonType int    `json:"person_type"`
	Name       string `json:"name"`
	Gender     *int   `json:"gender"`
	Mobile     string `json:"mobile"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Status     int    `json:"status"`

	StudentExtend *StudentExtend `json:"student_extend,omitempty"`
	StaffExtend   *StaffExtend   `json:"staff_extend,omitempty"`
}

// StudentExtend 学生扩展信息
type StudentExtend struct {
	PersonID         int    `json:"person_id" db:"person_id"`
	AreaID           *int   `json:"area_id" db:"area_id"`
	StudentNo        string `json:"student_no" db:"student_no"`
	Grade            string `json:"grade" db:"grade"`
	EducationLevel   string `json:"education_level" db:"education_level"`
	SchoolSystem     string `json:"school_system" db:"school_system"`
	IDCard           string `json:"id_card" db:"id_card"`
	AdmissionNo      string `json:"admission_no" db:"admission_no"`
	ExamNo           string `json:"exam_no" db:"exam_no"`
	EnrollmentStatus int    `json:"enrollment_status" db:"enrollment_status"`
	IsEnrolled       int    `json:"is_enrolled" db:"is_enrolled"`
	CollegeID        int    `json:"college_id" db:"college_id"`
	FacultyID        *int   `json:"faculty_id" db:"faculty_id"`
	ProfessionID     int    `json:"profession_id" db:"profession_id"`
	ClassID          *int   `json:"class_id" db:"class_id"`
}

// StaffExtend 政工扩展信息
type StaffExtend struct {
	PersonID     int    `json:"person_id" db:"person_id"`
	StaffNo      string `json:"staff_no" db:"staff_no"`
	DepartmentID *int   `json:"department_id" db:"department_id"`
	CollegeID    *int   `json:"college_id" db:"college_id"`
	FacultyID    *int   `json:"faculty_id" db:"faculty_id"`
}

// PersonListResponse 人员列表响应
type PersonListResponse struct {
	Items    []PersonWithExtend `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// PersonRole 人员角色关系
type PersonRole struct {
	CustomerID int `json:"customer_id" db:"customer_id"`
	PersonID   int `json:"person_id" db:"person_id"`
	RoleID     int `json:"role_id" db:"role_id"`
}

// PersonsRole 管辖角色
type PersonsRole struct {
	ID          int    `json:"id" db:"id"`
	CustomerID  int    `json:"customer_id" db:"customer_id"`
	ParentID    int    `json:"parent_id" db:"parent_id"`
	Name        string `json:"name" db:"name"`
	Permissions string `json:"permissions" db:"permissions"` // 菜单权限，逗号分隔的ID
}

// PersonHasDepartment 人员角色管辖机构关系
type PersonHasDepartment struct {
	CustomerID     int `json:"customer_id" db:"customer_id"`
	PersonsRolesID int `json:"persons_roles_id" db:"persons_roles_id"`
	PersonID       int `json:"person_id" db:"person_id"`
	DepartmentID   int `json:"department_id" db:"department_id"`
}
