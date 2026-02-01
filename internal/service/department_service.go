package service

import (
	"college_api/internal/model"
	"college_api/internal/repository"
	"fmt"
)

// DepartmentService 机构业务逻辑层
type DepartmentService struct {
	repo *repository.DepartmentRepository
}

// NewDepartmentService 创建机构服务
func NewDepartmentService(repo *repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

// GetDepartmentTree 获取机构树
func (s *DepartmentService) GetDepartmentTree(customerID int, visibleDeptIDs []int) ([]model.DepartmentNode, error) {
	println("[DEBUG] GetDepartmentTree: customerID =", customerID, ", visibleDeptIDs 数量 =", len(visibleDeptIDs))
	
	// 如果 visibleDeptIDs 为空（政工人员没有任何权限），返回空树
	if visibleDeptIDs != nil && len(visibleDeptIDs) == 0 {
		println("[DEBUG] GetDepartmentTree: 没有可见机构权限，返回空树")
		return []model.DepartmentNode{}, nil
	}

	// 1. 查询学校（第一级）
	println("[DEBUG] GetDepartmentTree: 查询学校节点")
	school, err := s.repo.GetSchool(customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get school: %w", err)
	}
	println("[DEBUG] GetDepartmentTree: 学校节点 - ID:", school.ID, ", 名称:", school.DepartmentName)

	// 2. 查询行政机构
	println("[DEBUG] GetDepartmentTree: 查询行政机构")
	adminDepts, err := s.repo.GetAdminDepartments(customerID, visibleDeptIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin departments: %w", err)
	}
	println("[DEBUG] GetDepartmentTree: 行政机构数量 =", len(adminDepts))

	// 3. 查询组织机构
	println("[DEBUG] GetDepartmentTree: 查询组织机构")
	orgDepts, err := s.repo.GetOrgDepartments(customerID, visibleDeptIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get org departments: %w", err)
	}
	println("[DEBUG] GetDepartmentTree: 组织机构数量 =", len(orgDepts))

	// 4. 合并所有机构
	allDepts := append(adminDepts, orgDepts...)
	println("[DEBUG] GetDepartmentTree: 合并后机构总数 =", len(allDepts))

	// 5. 构建树形结构
	tree := s.buildTreeStructure(school, allDepts)

	return tree, nil
}

// GetDepartmentList 获取机构列表
func (s *DepartmentService) GetDepartmentList(customerID int, keyword string, departmentType *int, visibleDeptIDs []int) ([]model.Department, error) {
	// 如果 visibleDeptIDs 为空（政工人员没有任何权限），返回空列表
	if visibleDeptIDs != nil && len(visibleDeptIDs) == 0 {
		println("[DEBUG] GetDepartmentList: 没有可见机构权限，返回空列表")
		return []model.Department{}, nil
	}

	return s.repo.SearchDepartments(customerID, keyword, departmentType, visibleDeptIDs)
}

// GetStaffVisibleDepartmentIDs 获取政工人员可见的机构ID列表（包含子机构）
func (s *DepartmentService) GetStaffVisibleDepartmentIDs(customerID, personID int) ([]int, error) {
	println("[DEBUG] GetStaffVisibleDepartmentIDs: customerID =", customerID, ", personID =", personID)
	
	// 1. 获取政工人员的所有角色
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 查询角色ID")
	roleIDs, err := s.repo.GetStaffRoleIDs(personID)
	if err != nil {
		println("[ERROR] GetStaffVisibleDepartmentIDs: 查询角色ID失败:", err.Error())
		return nil, fmt.Errorf("failed to get staff role ids: %w", err)
	}
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 角色ID数量 =", len(roleIDs))

	// 2. 合并两个来源的部门权限
	authorizedDeptIDMap := make(map[int]bool)

	// 2.1 获取角色关联的部门权限
	if len(roleIDs) > 0 {
		println("[DEBUG] GetStaffVisibleDepartmentIDs: 查询角色关联的部门权限")
		roleDeptIDs, err := s.repo.GetRoleDepartmentIDs(customerID, roleIDs)
		if err != nil {
			println("[ERROR] GetStaffVisibleDepartmentIDs: 查询角色部门权限失败:", err.Error())
			return nil, fmt.Errorf("failed to get role department ids: %w", err)
		}
		println("[DEBUG] GetStaffVisibleDepartmentIDs: 角色部门权限数量 =", len(roleDeptIDs))

		for _, id := range roleDeptIDs {
			authorizedDeptIDMap[id] = true
		}
	}

	// 2.2 获取直接分配给人员的部门权限
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 查询直接分配的部门权限")
	directDeptIDs, err := s.repo.GetDirectDepartmentIDs(personID)
	if err != nil {
		println("[ERROR] GetStaffVisibleDepartmentIDs: 查询直接部门权限失败:", err.Error())
		return nil, fmt.Errorf("failed to get direct department ids: %w", err)
	}
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 直接部门权限数量 =", len(directDeptIDs))

	for _, id := range directDeptIDs {
		authorizedDeptIDMap[id] = true
	}

	// 如果没有任何权限，返回空列表
	if len(authorizedDeptIDMap) == 0 {
		println("[WARN] GetStaffVisibleDepartmentIDs: 没有任何部门权限")
		return []int{}, nil
	}

	// 3. 转换为切片
	authorizedDeptIDs := make([]int, 0, len(authorizedDeptIDMap))
	for id := range authorizedDeptIDMap {
		authorizedDeptIDs = append(authorizedDeptIDs, id)
	}
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 合并后的部门权限数量 =", len(authorizedDeptIDs))

	// 4. 扩展为包含所有子部门的ID列表
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 扩展子部门")
	visibleDeptIDs, err := s.repo.ExpandDepartmentIDs(customerID, authorizedDeptIDs)
	if err != nil {
		println("[ERROR] GetStaffVisibleDepartmentIDs: 扩展子部门失败:", err.Error())
		return nil, fmt.Errorf("failed to expand department ids: %w", err)
	}
	println("[DEBUG] GetStaffVisibleDepartmentIDs: 扩展后的部门ID数量 =", len(visibleDeptIDs))

	return visibleDeptIDs, nil
}

// buildTreeStructure 构建树形结构
func (s *DepartmentService) buildTreeStructure(school *model.Department, depts []model.Department) []model.DepartmentNode {
	println("[DEBUG] buildTreeStructure: 开始构建树形结构")
	println("[DEBUG] buildTreeStructure: 学校ID =", school.ID, ", 学校名称 =", school.DepartmentName)
	println("[DEBUG] buildTreeStructure: 机构总数 =", len(depts))
	
	// 创建机构映射表
	deptMap := make(map[int]model.Department)
	for _, dept := range depts {
		deptMap[dept.ID] = dept
	}
	
	// 递归构建子树
	var buildNode func(dept model.Department) model.DepartmentNode
	buildNode = func(dept model.Department) model.DepartmentNode {
		node := model.DepartmentNode{
			ID:             dept.ID,
			ParentID:       dept.ParentID,
			RecommendNum:   dept.RecommendNum,
			DepartmentName: dept.DepartmentName,
			DepartmentType: dept.DepartmentType,
			TreeLevel:      dept.TreeLevel,
			Items:          []model.DepartmentNode{},
		}
		
		// 查找所有子节点
		for _, childDept := range depts {
			if childDept.ParentID == dept.ID {
				childNode := buildNode(childDept)
				node.Items = append(node.Items, childNode)
				println("[DEBUG] buildTreeStructure: 将机构", childDept.ID, "挂在父节点", dept.ID, "下")
			}
		}
		
		return node
	}
	
	// 初始化学校节点
	schoolNode := model.DepartmentNode{
		ID:             school.ID,
		ParentID:       school.ParentID,
		RecommendNum:   school.RecommendNum,
		DepartmentName: school.DepartmentName,
		DepartmentType: school.DepartmentType,
		TreeLevel:      school.TreeLevel,
		Items:          []model.DepartmentNode{},
	}
	
	// 查找所有 tree_level = 3 的机构作为学校的直接子节点
	level3Count := 0
	for _, dept := range depts {
		if dept.TreeLevel == 3 {
			childNode := buildNode(dept)
			schoolNode.Items = append(schoolNode.Items, childNode)
			level3Count++
			println("[DEBUG] buildTreeStructure: 将机构", dept.ID, "挂在学校节点下，子节点数:", len(childNode.Items))
		}
	}
	
	println("[DEBUG] buildTreeStructure: 学校节点下有", level3Count, "个直接子节点")
	println("[DEBUG] buildTreeStructure: 学校节点的 Items 数量 =", len(schoolNode.Items))

	return []model.DepartmentNode{schoolNode}
}
