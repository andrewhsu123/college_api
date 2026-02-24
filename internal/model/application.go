package model

// Application 应用模型
type Application struct {
	ID          int    `json:"id" db:"id"`
	CustomerID  int    `json:"customer_id" db:"customer_id"`
	AppCode     string `json:"app_code" db:"app_code"`
	AppName     string `json:"app_name" db:"app_name"`
	AppURL      string `json:"app_url" db:"app_url"`
	ShortIntro  string `json:"short_intro" db:"short_intro"`
	Description string `json:"description" db:"description"`
	Icon        string `json:"icon" db:"icon"`
	DataType    int    `json:"data_type" db:"data_type"` // 1=可见所有人数据 2=管辖范围下数据
	Status      int    `json:"status" db:"status"`       // 1=已启用 2=已停用
	CreatedAt   int    `json:"created_at" db:"created_at"`
}

// ApplicationScope 应用使用范围
type ApplicationScope struct {
	ID            int    `json:"id" db:"id"`
	CustomerID    int    `json:"customer_id" db:"customer_id"`
	ApplicationID int    `json:"application_id" db:"application_id"`
	RangeType     string `json:"range_type" db:"range_type"` // 1=全部人员 2=学生 3=政工 4=角色 5=机构
	StudentIDs    string `json:"student_ids" db:"student_ids"`
	StaffIDs      string `json:"staff_ids" db:"staff_ids"`
	RoleIDs       string `json:"role_ids" db:"role_ids"`
	DepartmentIDs string `json:"department_ids" db:"department_ids"`
}

// ApplicationListRequest 应用列表查询请求
type ApplicationListRequest struct {
	CustomerID int    `form:"customer_id" binding:"required"`
	AppCode    string `form:"app_code"`
	AppName    string `form:"app_name"`
	ShortIntro string `form:"short_intro"`
	DataType   *int   `form:"data_type"`
	Status     *int   `form:"status"`
}

// ApplicationVisibleRequest 可见应用查询请求
type ApplicationVisibleRequest struct {
	CustomerID    int    `form:"customer_id" binding:"required"`
	PersonID      *int   `form:"person_id"`
	PersonType    *int   `form:"person_type"`    // 1=学生 2=政工
	RoleIDs       string `form:"role_ids"`       // 用户拥有的角色ID，逗号分隔
	DepartmentIDs string `form:"department_ids"` // 用户管辖的机构ID，逗号分隔
}

// ApplicationWithScope 应用及其使用范围
type ApplicationWithScope struct {
	Application
	Scope *ApplicationScope `json:"scope,omitempty"`
}
