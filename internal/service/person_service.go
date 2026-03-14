package service

import (
	"college_api/internal/model"
	"college_api/internal/repository"
	"fmt"
	"strconv"
	"strings"
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

// GetStaffInfo 获取人员完整信息（支持学生、政工、维修工）
func (s *PersonService) GetStaffInfo(personID int) (model.PersonInfo, error) {
	info, err := s.repo.GetPersonInfo(personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person info: %w", err)
	}
	// 填充机构名称
	s.fillDepartmentNames(info)
	// 获取管辖角色和机构信息
	managedRoles, err := s.getManagedRoles(info.GetUniversityID(), personID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get managed roles for person %d: %v\n", personID, err)
		managedRoles = []model.ManagedRole{}
	}
	info.SetManagedRoles(managedRoles)

	// 获取菜单权限
	managedMenu, err := s.getManagedMenu(info.GetUniversityID(), personID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get managed menu for person %d: %v\n", personID, err)
		managedMenu = []int{}
	}
	info.SetManagedMenu(managedMenu)

	// 获取我的机构列表
	selfDepartment := s.getSelfDepartment(info)
	info.SetSelfDepartment(selfDepartment)

	// 获取我的角色列表
	selfRoles, err := s.deptRepo.GetStaffRoleIDs(info.GetUniversityID(), personID)
	if err != nil {
		fmt.Printf("[WARN] Failed to get self roles for person %d: %v\n", personID, err)
		selfRoles = []int{}
	}
	info.SetSelfRoles(selfRoles)

	return info, nil
}

// fillDepartmentNames 填充机构名称
func (s *PersonService) fillDepartmentNames(info model.PersonInfo) {
	// 获取学校名称
	if info.GetUniversityID() > 0 {
		name, err := s.repo.GetDepartmentName(info.GetUniversityID())
		if err == nil {
			info.SetUniversityName(name)
		}
	}

	// 根据具体类型填充不同的机构名称
	switch v := info.(type) {
	case *model.StudentPersonInfo:
		s.fillStudentDepartmentNames(v)
	case *model.StaffPersonInfo:
		s.fillStaffDepartmentNames(v)
	}
}

// getSelfDepartment 获取我的机构列表
func (s *PersonService) getSelfDepartment(info model.PersonInfo) []int {
	var deptIDs []int

	switch v := info.(type) {
	case *model.StaffPersonInfo:
		// 政工: department_id, college_id, faculty_id
		if v.DepartmentID != nil && *v.DepartmentID > 0 {
			deptIDs = append(deptIDs, *v.DepartmentID)
		}
		if v.CollegeID != nil && *v.CollegeID > 0 {
			deptIDs = append(deptIDs, *v.CollegeID)
		}
		if v.FacultyID != nil && *v.FacultyID > 0 {
			deptIDs = append(deptIDs, *v.FacultyID)
		}
	case *model.StudentPersonInfo:
		// 学生: college_id, faculty_id, profession_id, class_id
		if v.CollegeID != nil && *v.CollegeID > 0 {
			deptIDs = append(deptIDs, *v.CollegeID)
		}
		if v.FacultyID != nil && *v.FacultyID > 0 {
			deptIDs = append(deptIDs, *v.FacultyID)
		}
		if v.ProfessionID != nil && *v.ProfessionID > 0 {
			deptIDs = append(deptIDs, *v.ProfessionID)
		}
		if v.ClassID != nil && *v.ClassID > 0 {
			deptIDs = append(deptIDs, *v.ClassID)
		}
	}

	return deptIDs
}

// fillStudentDepartmentNames 填充学生机构名称
func (s *PersonService) fillStudentDepartmentNames(info *model.StudentPersonInfo) {
	if info.CollegeID != nil && *info.CollegeID > 0 {
		name, err := s.repo.GetDepartmentName(*info.CollegeID)
		if err == nil {
			info.CollegeName = &name
		}
	}
	if info.FacultyID != nil && *info.FacultyID > 0 {
		name, err := s.repo.GetDepartmentName(*info.FacultyID)
		if err == nil {
			info.FacultyName = &name
		}
	}
	if info.ProfessionID != nil && *info.ProfessionID > 0 {
		name, err := s.repo.GetDepartmentName(*info.ProfessionID)
		if err == nil {
			info.ProfessionName = &name
		}
	}
	if info.ClassID != nil && *info.ClassID > 0 {
		name, err := s.repo.GetDepartmentName(*info.ClassID)
		if err == nil {
			info.ClassName = &name
		}
	}
}

// fillStaffDepartmentNames 填充政工机构名称
func (s *PersonService) fillStaffDepartmentNames(info *model.StaffPersonInfo) {
	if info.DepartmentID != nil && *info.DepartmentID > 0 {
		name, err := s.repo.GetDepartmentName(*info.DepartmentID)
		if err == nil {
			info.DepartmentName = &name
		}
	}
	if info.CollegeID != nil && *info.CollegeID > 0 {
		name, err := s.repo.GetDepartmentName(*info.CollegeID)
		if err == nil {
			info.CollegeName = &name
		}
	}
	if info.FacultyID != nil && *info.FacultyID > 0 {
		name, err := s.repo.GetDepartmentName(*info.FacultyID)
		if err == nil {
			info.FacultyName = &name
		}
	}
}

// getManagedRoles 获取人员的管辖角色及机构列表
func (s *PersonService) getManagedRoles(customerID, personID int) ([]model.ManagedRole, error) {
	roles, err := s.deptRepo.GetPersonRoles(customerID, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person roles: %w", err)
	}

	if len(roles) == 0 {
		return []model.ManagedRole{}, nil
	}

	managedRoles := make([]model.ManagedRole, 0, len(roles))
	for _, role := range roles {
		parentName, _ := s.deptRepo.GetRoleParentName(customerID, role.ParentID)

		departments, err := s.deptRepo.GetRoleManagedDepartments(customerID, personID, role.ID)
		if err != nil {
			fmt.Printf("[WARN] Failed to get managed departments for role %d: %v\n", role.ID, err)
			departments = []model.ManagedDepartment{}
		}

		managedRoles = append(managedRoles, model.ManagedRole{
			ID:          role.ID,
			ParentID:    role.ParentID,
			ParentName:  parentName,
			Name:        role.Name,
			Departments: departments,
		})
	}

	return managedRoles, nil
}

// getManagedMenu 获取人员的菜单权限ID列表
func (s *PersonService) getManagedMenu(customerID, personID int) ([]int, error) {
	roles, err := s.deptRepo.GetPersonRoles(customerID, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person roles: %w", err)
	}

	if len(roles) == 0 {
		return []int{}, nil
	}

	menuIDMap := make(map[int]bool)
	for _, role := range roles {
		if role.Permissions == "" {
			continue
		}

		ids := strings.Split(role.Permissions, ",")
		for _, idStr := range ids {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			if id, err := strconv.Atoi(idStr); err == nil {
				menuIDMap[id] = true
			}
		}
	}

	menuIDs := make([]int, 0, len(menuIDMap))
	for id := range menuIDMap {
		menuIDs = append(menuIDs, id)
	}

	return menuIDs, nil
}

// GetManagedDepartmentIDs 获取人员管辖的所有机构ID（包含子机构，用于权限过滤）
func (s *PersonService) GetManagedDepartmentIDs(customerID, personID int) ([]int, error) {
	directDeptIDs, err := s.deptRepo.GetDirectDepartmentIDs(customerID, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct department ids: %w", err)
	}

	if len(directDeptIDs) == 0 {
		return []int{}, nil
	}

	visibleDeptIDs, err := s.deptRepo.ExpandDepartmentIDs(customerID, directDeptIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to expand department ids: %w", err)
	}

	return visibleDeptIDs, nil
}

// GetAdminUserInfo 获取学校管理员完整信息
func (s *PersonService) GetAdminUserInfo(userID int) (*model.AdminUserInfo, error) {
	info, err := s.repo.GetAdminUserInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user info: %w", err)
	}

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
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	persons, total, err := s.repo.GetPersonList(req, visibleDeptIDs, managedPersonIDs, isStaff)
	if err != nil {
		return nil, fmt.Errorf("failed to get person list: %w", err)
	}

	if req.WithExtend && len(persons) > 0 {
		personIDs := make([]int, len(persons))
		for i, p := range persons {
			personIDs[i] = p.ID
		}

		if req.PersonType == 1 {
			extendMap, err := s.repo.GetStudentExtendInfo(personIDs, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get student extend info: %w", err)
			}
			for i := range persons {
				if extend, ok := extendMap[persons[i].ID]; ok {
					persons[i].StudentExtend = extend
				}
			}
		} else if req.PersonType == 2 {
			extendMap, err := s.repo.GetStaffExtendInfo(personIDs, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get staff extend info: %w", err)
			}
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
