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
		WHERE customer_id = ? AND deleted_at = 0
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

// GetStaffRoleIDs 获取政工人员的角色ID列表
func (r *DepartmentRepository) GetStaffRoleIDs(personID int) ([]int, error) {
	query := `SELECT role_id FROM persons_has_roles WHERE person_id = ?`
	println("[SQL] GetStaffRoleIDs:", query, "| person_id =", personID)

	rows, err := r.db.Query(query, personID)
	if err != nil {
		println("[SQL ERROR] GetStaffRoleIDs:", err.Error())
		return nil, err
	}
	defer rows.Close()

	var roleIDs []int
	for rows.Next() {
		var roleID int
		if err := rows.Scan(&roleID); err != nil {
			println("[SQL ERROR] GetStaffRoleIDs Scan:", err.Error())
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	println("[SQL RESULT] GetStaffRoleIDs: 查询到", len(roleIDs), "个角色")
	return roleIDs, nil
}

// GetRoleDepartmentIDs 获取角色的部门权限ID列表
func (r *DepartmentRepository) GetRoleDepartmentIDs(customerID int, roleIDs []int) ([]int, error) {
	if len(roleIDs) == 0 {
		return []int{}, nil
	}

	placeholders := make([]string, len(roleIDs))
	args := []interface{}{customerID}
	for i, id := range roleIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT department_ids 
		FROM persons_roles 
		WHERE customer_id = ? AND id IN (%s)
	`, strings.Join(placeholders, ","))
	
	println("[SQL] GetRoleDepartmentIDs:", query)
	println("[SQL PARAMS] customer_id =", customerID, ", role_ids =", roleIDs)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		println("[SQL ERROR] GetRoleDepartmentIDs:", err.Error())
		return nil, err
	}
	defer rows.Close()

	deptIDMap := make(map[int]bool)
	rowCount := 0
	for rows.Next() {
		rowCount++
		var departmentIDs string
		if err := rows.Scan(&departmentIDs); err != nil {
			println("[SQL ERROR] GetRoleDepartmentIDs Scan:", err.Error())
			return nil, err
		}

		println("[SQL ROW]", rowCount, "department_ids =", departmentIDs)

		// 解析 JSON 数组格式的 department_ids
		// 简单处理：去掉 [ ] 和空格，按逗号分割
		departmentIDs = strings.Trim(departmentIDs, "[]")
		if departmentIDs == "" {
			continue
		}

		ids := strings.Split(departmentIDs, ",")
		for _, idStr := range ids {
			var id int
			if _, err := fmt.Sscanf(strings.TrimSpace(idStr), "%d", &id); err == nil {
				deptIDMap[id] = true
			}
		}
	}

	result := make([]int, 0, len(deptIDMap))
	for id := range deptIDMap {
		result = append(result, id)
	}

	println("[SQL RESULT] GetRoleDepartmentIDs: 查询到", rowCount, "行，解析出", len(result), "个部门ID")
	return result, nil
}

// GetRolePersonIDs 获取角色的人员权限ID列表
func (r *DepartmentRepository) GetRolePersonIDs(customerID int, roleIDs []int) ([]int, error) {
	if len(roleIDs) == 0 {
		return []int{}, nil
	}

	placeholders := make([]string, len(roleIDs))
	args := []interface{}{customerID}
	for i, id := range roleIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT person_ids 
		FROM persons_roles 
		WHERE customer_id = ? AND id IN (%s)
	`, strings.Join(placeholders, ","))

	println("[SQL] GetRolePersonIDs:", query)
	println("[SQL PARAMS] customer_id =", customerID, ", role_ids =", roleIDs)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		println("[SQL ERROR] GetRolePersonIDs:", err.Error())
		return nil, err
	}
	defer rows.Close()

	personIDMap := make(map[int]bool)
	rowCount := 0
	for rows.Next() {
		rowCount++
		var personIDs string
		if err := rows.Scan(&personIDs); err != nil {
			println("[SQL ERROR] GetRolePersonIDs Scan:", err.Error())
			return nil, err
		}

		println("[SQL ROW]", rowCount, "person_ids =", personIDs)

		// 解析 JSON 数组格式的 person_ids
		// 简单处理：去掉 [ ] 和空格，按逗号分割
		personIDs = strings.Trim(personIDs, "[]")
		if personIDs == "" {
			continue
		}

		ids := strings.Split(personIDs, ",")
		for _, idStr := range ids {
			var id int
			if _, err := fmt.Sscanf(strings.TrimSpace(idStr), "%d", &id); err == nil {
				personIDMap[id] = true
			}
		}
	}

	result := make([]int, 0, len(personIDMap))
	for id := range personIDMap {
		result = append(result, id)
	}

	println("[SQL RESULT] GetRolePersonIDs: 查询到", rowCount, "行，解析出", len(result), "个人员ID")
	return result, nil
}
func (r *DepartmentRepository) GetDirectDepartmentIDs(personID int) ([]int, error) {
	query := `SELECT department_id FROM persons_has_department WHERE person_id = ?`
	println("[SQL] GetDirectDepartmentIDs:", query, "| person_id =", personID)

	rows, err := r.db.Query(query, personID)
	if err != nil {
		println("[SQL ERROR] GetDirectDepartmentIDs:", err.Error())
		return nil, err
	}
	defer rows.Close()

	var deptIDs []int
	for rows.Next() {
		var deptID int
		if err := rows.Scan(&deptID); err != nil {
			println("[SQL ERROR] GetDirectDepartmentIDs Scan:", err.Error())
			return nil, err
		}
		deptIDs = append(deptIDs, deptID)
	}

	println("[SQL RESULT] GetDirectDepartmentIDs: 查询到", len(deptIDs), "个直接部门权限")
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
