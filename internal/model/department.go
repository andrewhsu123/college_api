package model

import "time"

// Department 机构模型
type Department struct {
	ID             int       `json:"id" db:"id"`
	CustomerID     int       `json:"customer_id" db:"customer_id"`
	ParentID       int       `json:"parent_id" db:"parent_id"`
	DepartmentName string    `json:"department_name" db:"department_name"`
	DepartmentType int       `json:"department_type" db:"department_type"` // 0=学校 1=行政机构 2=学院 3=系 4=专业 5=班级
	RecommendNum   int       `json:"recommend_num" db:"recommend_num"`     // 下级机构数量
	TreeLeft       int       `json:"-" db:"tree_left"`                     // 不在 JSON 中显示
	TreeRight      int       `json:"-" db:"tree_right"`                    // 不在 JSON 中显示
	TreeLevel      int       `json:"-" db:"tree_level"`                    // 不在 JSON 中显示
	Sort           int       `json:"-" db:"sort"`                          // 不在 JSON 中显示
	CreatedAt      time.Time `json:"-" db:"created_at"`                    // 不在 JSON 中显示
	UpdatedAt      time.Time `json:"-" db:"updated_at"`                    // 不在 JSON 中显示
	DeletedAt      int       `json:"-" db:"deleted_at"`                    // 不在 JSON 中显示
}

// DepartmentNode 机构树节点
type DepartmentNode struct {
	ID             int              `json:"id"`
	ParentID       int              `json:"parent_id"`
	RecommendNum   int              `json:"recommend_num"`
	DepartmentName string           `json:"department_name"`
	DepartmentType int              `json:"department_type"`
	TreeLevel      int              `json:"tree_level"`
	Items          []DepartmentNode `json:"items"`
}
