package repository

import (
	"college_api/internal/model"
	"database/sql"
	"fmt"
	"strings"
)

// DepartmentRepository 机构数据访问层
type DepartmentRepository struct {
	db *sql.DB
}

// NewDepartmentRepository 创建机构仓库
func NewDepartmentRepository(db *sql.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// GetSchool 查询学校机构
func (r *DepartmentRepository) GetSchool(customerID int) (*model.Department, error) {
	query := `
        SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
        FROM departments
        WHERE customer_id = ? AND department_type = 0 AND deleted_at = 0
        LIMIT 1
    `
	var dept model.Department
	err := r.db.QueryRow(query, customerID).Scan(
		&dept.ID,
		&dept.ParentID,
		&dept.RecommendNum,
		&dept.DepartmentName,
		&dept.DepartmentType,
		&dept.TreeLevel,
	)
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// GetAdminDepartments 查询行政机构
func (r *DepartmentRepository) GetAdminDepartments(customerID int, visibleDeptIDs []int) ([]model.Department, error) {
	query := `
        SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
        FROM departments
        WHERE customer_id = ? 
          AND tree_level > 2 
          AND department_type = 1 
          AND deleted_at = 0
    `
	args := []interface{}{customerID}
	// 政工人员：添加权限过滤
	if visibleDeptIDs != nil && len(visibleDeptIDs) > 0 {
		placeholders := make([]string, len(visibleDeptIDs))
		for i, id := range visibleDeptIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}
	query += " ORDER BY tree_left"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDepartments(rows)
}

// GetOrgDepartments 查询组织机构
func (r *DepartmentRepository) GetOrgDepartments(customerID int, visibleDeptIDs []int) ([]model.Department, error) {
	query := `
        SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
        FROM departments
        WHERE customer_id = ? 
            AND tree_level > 2 
            AND department_type != 1 
            AND deleted_at = 0
    `
	args := []interface{}{customerID}

	// 政工人员：添加权限过滤
	if visibleDeptIDs != nil && len(visibleDeptIDs) > 0 {
		placeholders := make([]string, len(visibleDeptIDs))
		for i, id := range visibleDeptIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}

	query += " ORDER BY tree_left"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDepartments(rows)
}

// SearchDepartments 搜索机构列表
func (r *DepartmentRepository) SearchDepartments(customerID int, keyword string, departmentType *int, visibleDeptIDs []int) ([]model.Department, error) {
	query := `
        SELECT id, parent_id, recommend_num, department_name, department_type, tree_level
        FROM departments
        WHERE customer_id = ? AND deleted_at = 0 AND tree_level IS NOT NULL
    `

	args := []interface{}{customerID}
	// 名称模糊查询
	if keyword != "" {
		query += " AND department_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	// 类型查询
	if departmentType != nil {
		query += " AND department_type = ?"
		args = append(args, *departmentType)
	}
	// 政工人员：添加权限过滤
	if visibleDeptIDs != nil && len(visibleDeptIDs) > 0 {
		placeholders := make([]string, len(visibleDeptIDs))
		for i, id := range visibleDeptIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}
	query += " ORDER BY tree_level, tree_left"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanDepartments(rows)
}

// GetStaffRoleIDs 获取人员的角色ID列表（新表结构：先查customer_id再查role_id）
func (r *DepartmentRepository) GetStaffRoleIDs(customerID, personID int) ([]int, error) {
	query := `SELECT role_id FROM persons_has_roles WHERE customer_id = ? AND person_id = ?`

	rows, err := r.db.Query(query, customerID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []int
	for rows.Next() {
		var roleID int
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, nil
}

// GetPersonRoles 获取人员的角色详情列表
func (r *DepartmentRepository) GetPersonRoles(customerID, personID int) ([]model.PersonsRole, error) {
	query := `
        SELECT pr.id, pr.customer_id, pr.parent_id, pr.name, COALESCE(pr.permissions, '') as permissions
        FROM persons_roles pr
        INNER JOIN persons_has_roles phr ON pr.id = phr.role_id AND pr.customer_id = phr.customer_id
        WHERE phr.customer_id = ? AND phr.person_id = ? AND pr.deleted_at = 0
    `

	rows, err := r.db.Query(query, customerID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.PersonsRole
	for rows.Next() {
		var role model.PersonsRole
		if err := rows.Scan(&role.ID, &role.CustomerID, &role.ParentID, &role.Name, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoleParentName 获取角色的上级角色组名称
func (r *DepartmentRepository) GetRoleParentName(customerID, parentID int) (string, error) {
	if parentID == 0 {
		return "", nil
	}

	query := `SELECT name FROM persons_roles WHERE customer_id = ? AND id = ? AND deleted_at = 0 LIMIT 1`

	var name string
	err := r.db.QueryRow(query, customerID, parentID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return name, nil
}

// GetRoleManagedDepartments 获取角色+人员对应的管辖机构列表
func (r *DepartmentRepository) GetRoleManagedDepartments(customerID, personID, roleID int) ([]model.ManagedDepartment, error) {
	query := `
        SELECT d.id, d.parent_id, d.department_name, d.department_type, 1 as status
        FROM departments d
        INNER JOIN persons_has_department phd ON d.id = phd.department_id
        WHERE phd.customer_id = ? AND phd.person_id = ? AND phd.persons_roles_id = ?
          AND d.deleted_at = 0
        ORDER BY d.tree_left
    `

	rows, err := r.db.Query(query, customerID, personID, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []model.ManagedDepartment
	for rows.Next() {
		var dept model.ManagedDepartment
		if err := rows.Scan(&dept.ID, &dept.ParentID, &dept.DepartmentName, &dept.DepartmentType, &dept.Status); err != nil {
			return nil, err
		}
		departments = append(departments, dept)
	}

	return departments, nil
}

// GetDirectDepartmentIDs 获取人员在某角色下的直接管辖机构ID列表
func (r *DepartmentRepository) GetDirectDepartmentIDs(customerID, personID int) ([]int, error) {
	query := `SELECT department_id FROM persons_has_department WHERE customer_id = ? AND person_id = ?`

	rows, err := r.db.Query(query, customerID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deptIDs []int
	for rows.Next() {
		var deptID int
		if err := rows.Scan(&deptID); err != nil {
			return nil, err
		}
		deptIDs = append(deptIDs, deptID)
	}

	return deptIDs, nil
}

// ExpandDepartmentIDs 扩展部门ID列表（包含所有子部门）
func (r *DepartmentRepository) ExpandDepartmentIDs(customerID int, deptIDs []int) ([]int, error) {
	if len(deptIDs) == 0 {
		return []int{}, nil
	}

	placeholders := make([]string, len(deptIDs))
	args := []interface{}{customerID}
	for i, id := range deptIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
        SELECT DISTINCT d.id
        FROM departments d
        WHERE d.customer_id = ?
          AND d.deleted_at = 0
          AND EXISTS (
            SELECT 1 
            FROM departments p
            WHERE p.id IN (%s)
              AND d.tree_left >= p.tree_left
              AND d.tree_right <= p.tree_right
          )
    `, strings.Join(placeholders, ","))

	println("[SQL] ExpandDepartmentIDs:", query)
	println("[SQL PARAMS] customer_id =", customerID, ", dept_ids =", deptIDs)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		println("[SQL ERROR] ExpandDepartmentIDs:", err.Error())
		return nil, err
	}
	defer rows.Close()

	var result []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			println("[SQL ERROR] ExpandDepartmentIDs Scan:", err.Error())
			return nil, err
		}
		result = append(result, id)
	}

	println("[SQL RESULT] ExpandDepartmentIDs: 扩展后共", len(result), "个部门ID")
	return result, nil
}

// scanDepartments 扫描部门列表
func (r *DepartmentRepository) scanDepartments(rows *sql.Rows) ([]model.Department, error) {
	var departments []model.Department

	for rows.Next() {
		var dept model.Department
		err := rows.Scan(
			&dept.ID,
			&dept.ParentID,
			&dept.RecommendNum,
			&dept.DepartmentName,
			&dept.DepartmentType,
			&dept.TreeLevel,
		)
		if err != nil {
			return nil, err
		}
		departments = append(departments, dept)
	}

	return departments, nil
}
