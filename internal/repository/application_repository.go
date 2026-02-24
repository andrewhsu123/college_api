package repository

import (
	"college_api/internal/model"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// ApplicationRepository 应用数据访问层
type ApplicationRepository struct {
	db *sql.DB
}

// NewApplicationRepository 创建应用仓库
func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

// GetApplicationList 查询应用列表
func (r *ApplicationRepository) GetApplicationList(req *model.ApplicationListRequest) ([]model.Application, error) {
	query := `
        SELECT id, customer_id, app_code, app_name, 
               COALESCE(app_url, '') as app_url,
               COALESCE(short_intro, '') as short_intro,
               COALESCE(description, '') as description,
               COALESCE(icon, '') as icon,
               data_type, status, created_at
        FROM applications
        WHERE customer_id = ? AND deleted_at = 0
    `
	args := []interface{}{req.CustomerID}

	if req.AppCode != "" {
		query += " AND app_code LIKE ?"
		args = append(args, "%"+req.AppCode+"%")
	}
	if req.AppName != "" {
		query += " AND app_name LIKE ?"
		args = append(args, "%"+req.AppName+"%")
	}
	if req.ShortIntro != "" {
		query += " AND short_intro LIKE ?"
		args = append(args, "%"+req.ShortIntro+"%")
	}
	if req.DataType != nil {
		query += " AND data_type = ?"
		args = append(args, *req.DataType)
	}
	if req.Status != nil {
		query += " AND status = ?"
		args = append(args, *req.Status)
	}

	query += " ORDER BY id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		err := rows.Scan(
			&app.ID, &app.CustomerID, &app.AppCode, &app.AppName,
			&app.AppURL, &app.ShortIntro, &app.Description, &app.Icon,
			&app.DataType, &app.Status, &app.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, nil
}

// GetApplicationScopes 获取应用的使用范围
func (r *ApplicationRepository) GetApplicationScopes(customerID int, appIDs []int) (map[int]*model.ApplicationScope, error) {
	if len(appIDs) == 0 {
		return make(map[int]*model.ApplicationScope), nil
	}

	placeholders := make([]string, len(appIDs))
	args := []interface{}{customerID}
	for i, id := range appIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
        SELECT id, customer_id, application_id, range_type,
               COALESCE(student_ids, '') as student_ids,
               COALESCE(staff_ids, '') as staff_ids,
               COALESCE(role_ids, '') as role_ids,
               COALESCE(department_ids, '') as department_ids
        FROM application_scopes
        WHERE customer_id = ? AND application_id IN (%s)
    `, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scopeMap := make(map[int]*model.ApplicationScope)
	for rows.Next() {
		var scope model.ApplicationScope
		err := rows.Scan(
			&scope.ID, &scope.CustomerID, &scope.ApplicationID, &scope.RangeType,
			&scope.StudentIDs, &scope.StaffIDs, &scope.RoleIDs, &scope.DepartmentIDs,
		)
		if err != nil {
			return nil, err
		}
		scopeMap[scope.ApplicationID] = &scope
	}

	return scopeMap, nil
}

// GetVisibleApplications 获取可见应用列表
func (r *ApplicationRepository) GetVisibleApplications(req *model.ApplicationVisibleRequest) ([]model.ApplicationWithScope, error) {
	// 从请求参数解析用户的角色和机构ID
	userRoleIDs := r.parseIDs(req.RoleIDs)
	userDeptIDs := r.parseIDs(req.DepartmentIDs)

	// 查询所有启用的应用
	query := `
        SELECT a.id, a.customer_id, a.app_code, a.app_name,
               COALESCE(a.app_url, '') as app_url,
               COALESCE(a.short_intro, '') as short_intro,
               COALESCE(a.description, '') as description,
               COALESCE(a.icon, '') as icon,
               a.data_type, a.status, a.created_at,
               COALESCE(s.id, 0) as scope_id,
               COALESCE(s.range_type, '1') as range_type,
               COALESCE(s.student_ids, '') as student_ids,
               COALESCE(s.staff_ids, '') as staff_ids,
               COALESCE(s.role_ids, '') as role_ids,
               COALESCE(s.department_ids, '') as department_ids
        FROM applications a
        LEFT JOIN application_scopes s ON a.id = s.application_id AND a.customer_id = s.customer_id
        WHERE a.customer_id = ? AND a.status = 1 AND a.deleted_at = 0
        ORDER BY a.id DESC
    `

	rows, err := r.db.Query(query, req.CustomerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.ApplicationWithScope
	for rows.Next() {
		var app model.Application
		var scopeID int
		var rangeType, studentIDs, staffIDs, roleIDs, departmentIDs string

		err := rows.Scan(
			&app.ID, &app.CustomerID, &app.AppCode, &app.AppName,
			&app.AppURL, &app.ShortIntro, &app.Description, &app.Icon,
			&app.DataType, &app.Status, &app.CreatedAt,
			&scopeID, &rangeType, &studentIDs, &staffIDs, &roleIDs, &departmentIDs,
		)
		if err != nil {
			return nil, err
		}

		// 检查是否可见
		visible := r.checkVisibility(req, rangeType, studentIDs, staffIDs, roleIDs, departmentIDs, userRoleIDs, userDeptIDs)
		if !visible {
			continue
		}

		item := model.ApplicationWithScope{Application: app}
		if scopeID > 0 {
			item.Scope = &model.ApplicationScope{
				ID:            scopeID,
				CustomerID:    req.CustomerID,
				ApplicationID: app.ID,
				RangeType:     rangeType,
				StudentIDs:    studentIDs,
				StaffIDs:      staffIDs,
				RoleIDs:       roleIDs,
				DepartmentIDs: departmentIDs,
			}
		}
		results = append(results, item)
	}

	return results, nil
}

// checkVisibility 检查应用是否对请求者可见
func (r *ApplicationRepository) checkVisibility(req *model.ApplicationVisibleRequest, rangeType, studentIDs, staffIDs, roleIDs, departmentIDs string, userRoleIDs, userDeptIDs []int) bool {
	// 解析 range_type（可能是逗号分隔的多个值）
	rangeTypes := strings.Split(rangeType, ",")
	rangeTypeMap := make(map[string]bool)
	for _, rt := range rangeTypes {
		rangeTypeMap[strings.TrimSpace(rt)] = true
	}

	// 1=全部人员，直接可见
	if rangeTypeMap["1"] {
		return true
	}

	// 根据 person_type 检查
	if req.PersonID != nil && req.PersonType != nil {
		// 2=学生范围
		if *req.PersonType == 1 && rangeTypeMap["2"] {
			if r.containsID(studentIDs, *req.PersonID) {
				return true
			}
		}
		// 3=政工范围
		if *req.PersonType == 2 && rangeTypeMap["3"] {
			if r.containsID(staffIDs, *req.PersonID) {
				return true
			}
		}
	}

	// 4=角色范围：检查用户是否拥有 role_ids 中的任一角色
	if rangeTypeMap["4"] && len(userRoleIDs) > 0 {
		scopeRoleIDs := r.parseIDs(roleIDs)
		for _, userRoleID := range userRoleIDs {
			for _, scopeRoleID := range scopeRoleIDs {
				if userRoleID == scopeRoleID {
					return true
				}
			}
		}
	}

	// 5=机构范围：检查用户管辖的机构是否与 department_ids 有交集
	if rangeTypeMap["5"] && len(userDeptIDs) > 0 {
		scopeDeptIDs := r.parseIDs(departmentIDs)
		for _, userDeptID := range userDeptIDs {
			for _, scopeDeptID := range scopeDeptIDs {
				if userDeptID == scopeDeptID {
					return true
				}
			}
		}
	}

	return false
}

// parseIDs 解析逗号分隔的ID字符串为整数切片
func (r *ApplicationRepository) parseIDs(idsStr string) []int {
	if idsStr == "" {
		return nil
	}

	var result []int
	ids := strings.Split(idsStr, ",")
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.Atoi(idStr); err == nil {
			result = append(result, id)
		}
	}
	return result
}

// containsID 检查逗号分隔的ID字符串中是否包含指定ID
func (r *ApplicationRepository) containsID(idsStr string, targetID int) bool {
	if idsStr == "" {
		return false
	}

	ids := strings.Split(idsStr, ",")
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.Atoi(idStr); err == nil && id == targetID {
			return true
		}
	}
	return false
}
