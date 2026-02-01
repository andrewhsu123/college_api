package service

import (
	"college_api/internal/model"
	"college_api/internal/repository"
	"fmt"
)

// PersonService 人员业务逻辑层
type PersonService struct {
	repo     *repository.PersonRepository
	deptRepo *repository.DepartmentRepository
}

// NewPersonService 创建人员服务
func NewPersonService(repo *repository.PersonRepository, deptRepo *repository.DepartmentRepository) *PersonService {
	return &PersonService{
		repo:     repo,
		deptRepo: deptRepo,
	}
}

// GetStaffInfo 获取政工完整信息
func (s *PersonService) GetStaffInfo(personID int) (*model.StaffInfo, error) {
	info, err := s.repo.GetStaffInfo(personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff info: %w", err)
	}

	// 获取机构名称
	if err := s.fillDepartmentNames(info); err != nil {
		fmt.Printf("[WARN] Failed to fill department names for person %d: %v\n", personID, err)
	}

	// 获取管辖的机构ID列表（包含子机构）
	managedDeptIDs, err := s.getManagedDepartmentIDs(info.UniversityID, personID)
	if err != nil {
		// 如果获取失败，记录错误但不影响主流程，返回空数组
		fmt.Printf("[WARN] Failed to get managed department IDs for person %d: %v\n", personID, err)
		info.ManagedDepartmentIDs = []int{}
	} else {
		info.ManagedDepartmentIDs = managedDeptIDs
	}

	// 获取管辖的人员ID列表
	managedPersonIDs, err := s.getManagedPersonIDs(info.UniversityID, personID)
	if err != nil {
		// 如果获取失败，记录错误但不影响主流程，返回空数组
		fmt.Printf("[WARN] Failed to get managed person IDs for person %d: %v\n", personID, err)
		info.ManagedPersonIDs = []int{}
	} else {
		info.ManagedPersonIDs = managedPersonIDs
	}

	return info, nil
}

// fillDepartmentNames 填充机构名称
func (s *PersonService) fillDepartmentNames(info *model.StaffInfo) error {
	// 获取大学名称
	if info.UniversityID > 0 {
		name, err := s.repo.GetDepartmentName(info.UniversityID)
		if err == nil {
			info.UniversityName = name
		}
	}

	// 获取部门名称
	if info.DepartmentID != nil && *info.DepartmentID > 0 {
		name, err := s.repo.GetDepartmentName(*info.DepartmentID)
		if err == nil {
			info.DepartmentName = &name
		}
	}

	// 获取学院名称
	if info.CollegeID != nil && *info.CollegeID > 0 {
		name, err := s.repo.GetDepartmentName(*info.CollegeID)
		if err == nil {
			info.CollegeName = &name
		}
	}

	// 获取系名称
	if info.FacultyID != nil && *info.FacultyID > 0 {
		name, err := s.repo.GetDepartmentName(*info.FacultyID)
		if err == nil {
			info.FacultyName = &name
		}
	}

	return nil
}

// getManagedDepartmentIDs 获取政工人员管辖的机构ID列表（包含子机构）
func (s *PersonService) getManagedDepartmentIDs(customerID, personID int) ([]int, error) {
	// 1. 获取政工人员的所有角色
	roleIDs, err := s.deptRepo.GetStaffRoleIDs(personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff role ids: %w", err)
	}
	// 2. 合并两个来源的部门权限
	authorizedDeptIDMap := make(map[int]bool)
	// 2.1 获取角色关联的部门权限
	if len(roleIDs) > 0 {
		roleDeptIDs, err := s.deptRepo.GetRoleDepartmentIDs(customerID, roleIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get role department ids: %w", err)
		}
		for _, id := range roleDeptIDs {
			authorizedDeptIDMap[id] = true
		}
	}
	// 2.2 获取直接分配给人员的部门权限
	directDeptIDs, err := s.deptRepo.GetDirectDepartmentIDs(personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct department ids: %w", err)
	}
	for _, id := range directDeptIDs {
		authorizedDeptIDMap[id] = true
	}
	// 如果没有任何权限，返回空列表
	if len(authorizedDeptIDMap) == 0 {
		return []int{}, nil
	}
	// 3. 转换为切片
	authorizedDeptIDs := make([]int, 0, len(authorizedDeptIDMap))
	for id := range authorizedDeptIDMap {
		authorizedDeptIDs = append(authorizedDeptIDs, id)
	}
	// 4. 扩展为包含所有子部门的ID列表
	visibleDeptIDs, err := s.deptRepo.ExpandDepartmentIDs(customerID, authorizedDeptIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to expand department ids: %w", err)
	}
	return visibleDeptIDs, nil
}

// getManagedPersonIDs 获取政工人员管辖的人员ID列表
func (s *PersonService) getManagedPersonIDs(customerID, personID int) ([]int, error) {
	// 1. 获取政工人员的所有角色
	roleIDs, err := s.deptRepo.GetStaffRoleIDs(personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff role ids: %w", err)
	}

	if len(roleIDs) == 0 {
		return []int{}, nil
	}

	// 2. 获取这些角色的人员权限
	personIDs, err := s.deptRepo.GetRolePersonIDs(customerID, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get role person ids: %w", err)
	}

	return personIDs, nil
}

// GetAdminUserInfo 获取学校管理员完整信息
func (s *PersonService) GetAdminUserInfo(userID int) (*model.AdminUserInfo, error) {
	info, err := s.repo.GetAdminUserInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user info: %w", err)
	}

	// 获取大学名称
	if info.UniversityID > 0 {
		name, err := s.repo.GetDepartmentName(info.UniversityID)
		if err == nil {
			info.UniversityName = name
		}
	}

	return info, nil
}

// GetPersonList 查询人员列表
func (s *PersonService) GetPersonList(req *model.PersonListRequest, isStaff bool, visibleDeptIDs, managedPersonIDs []int) (*model.PersonListResponse, error) {
	// 设置默认分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 查询人员列表
	persons, total, err := s.repo.GetPersonList(req, visibleDeptIDs, managedPersonIDs, isStaff)
	if err != nil {
		return nil, fmt.Errorf("failed to get person list: %w", err)
	}

	// 如果需要扩展信息，批量查询
	if req.WithExtend && len(persons) > 0 {
		personIDs := make([]int, len(persons))
		for i, p := range persons {
			personIDs[i] = p.ID
		}

		if req.PersonType == 1 {
			// 查询学生扩展信息
			extendMap, err := s.repo.GetStudentExtendInfo(personIDs, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get student extend info: %w", err)
			}

			// 合并扩展信息
			for i := range persons {
				if extend, ok := extendMap[persons[i].ID]; ok {
					persons[i].StudentExtend = extend
				}
			}
		} else if req.PersonType == 2 {
			// 查询政工扩展信息
			extendMap, err := s.repo.GetStaffExtendInfo(personIDs, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get staff extend info: %w", err)
			}

			// 合并扩展信息
			for i := range persons {
				if extend, ok := extendMap[persons[i].ID]; ok {
					persons[i].StaffExtend = extend
				}
			}
		}
	}

	return &model.PersonListResponse{
		Items:    persons,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
