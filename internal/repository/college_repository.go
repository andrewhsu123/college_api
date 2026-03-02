package repository

import (
	"college_api/internal/model"
	"database/sql"
	"fmt"
)

// CollegeRepository 学校数据访问层
type CollegeRepository struct {
	db *sql.DB
}

// NewCollegeRepository 创建学校仓库
func NewCollegeRepository(db *sql.DB) *CollegeRepository {
	return &CollegeRepository{db: db}
}

// GetCollegeList 查询学校列表
func (r *CollegeRepository) GetCollegeList(req *model.OpenCollegeRequest) (*model.OpenCollegeResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 1000 {
		req.PageSize = 1000
	}

	query := "SELECT customer_id, username, COALESCE(mobile, '') as mobile, status FROM admin_users WHERE creator_id = 0 AND deleted_at = 0"
	args := []any{}

	if req.Username != "" {
		query += " AND username LIKE ?"
		args = append(args, "%"+req.Username+"%")
	}
	if req.Mobile != "" {
		query += " AND mobile LIKE ?"
		args = append(args, "%"+req.Mobile+"%")
	}
	if req.Status != nil {
		query += " AND status = ?"
		args = append(args, *req.Status)
	}

	// count
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// pagination
	query += " ORDER BY customer_id ASC LIMIT ? OFFSET ?"
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var items []model.OpenCollegeItem
	for rows.Next() {
		var item model.OpenCollegeItem
		if err := rows.Scan(&item.CustomerID, &item.Username, &item.Mobile, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return &model.OpenCollegeResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetCampusAreaList 查询校区列表（level<=1：学校和校区）
func (r *CollegeRepository) GetCampusAreaList(req *model.OpenCampusAreaRequest) ([]model.OpenCampusAreaItem, error) {
	query := "SELECT id, area_name FROM campus_areas WHERE customer_id = ? AND level <= 1 AND deleted_at = 0 ORDER BY id ASC"
	rows, err := r.db.Query(query, req.UniversityID)
	if err != nil {
		return nil, fmt.Errorf("查询校区失败: %w", err)
	}
	defer rows.Close()

	var items []model.OpenCampusAreaItem
	for rows.Next() {
		var item model.OpenCampusAreaItem
		if err := rows.Scan(&item.ID, &item.AreaName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetDepartmentList 查询部门列表
func (r *CollegeRepository) GetDepartmentList(req *model.OpenDepartmentListRequest) ([]model.OpenDepartmentListItem, error) {
	query := "SELECT id, parent_id, department_name, department_type FROM departments WHERE customer_id = ? AND deleted_at = 0 ORDER BY sort ASC, id ASC"
	rows, err := r.db.Query(query, req.UniversityID)
	if err != nil {
		return nil, fmt.Errorf("查询部门列表失败: %w", err)
	}
	defer rows.Close()

	var items []model.OpenDepartmentListItem
	for rows.Next() {
		var item model.OpenDepartmentListItem
		if err := rows.Scan(&item.ID, &item.ParentID, &item.DepartmentName, &item.DepartmentType); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
