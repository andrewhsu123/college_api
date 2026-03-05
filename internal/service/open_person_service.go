package service

import (
	"college_api/internal/model"
	"college_api/internal/repository"
)

// OpenPersonService 开放接口人员服务
type OpenPersonService struct {
	repo *repository.OpenPersonRepository
}

// NewOpenPersonService 创建开放接口人员服务
func NewOpenPersonService(repo *repository.OpenPersonRepository) *OpenPersonService {
	return &OpenPersonService{repo: repo}
}

// GetStaffList 查询政工列表
func (s *OpenPersonService) GetStaffList(req *model.OpenStaffRequest) (*model.OpenStaffResponse, error) {
	return s.repo.GetStaffList(req)
}

// GetStudentList 查询学生列表
func (s *OpenPersonService) GetStudentList(req *model.OpenStudentRequest) (*model.OpenStudentResponse, error) {
	return s.repo.GetStudentList(req)
}

// GetRoleList 查询角色列表
func (s *OpenPersonService) GetRoleList(req *model.OpenRoleRequest) ([]model.OpenRoleItem, error) {
	return s.repo.GetRoleList(req)
}

// GetRolePersons 根据角色查询人员详情列表
func (s *OpenPersonService) GetRolePersons(req *model.OpenRolePersonsRequest) (*model.OpenRolePersonsResponse, error) {
	return s.repo.GetRolePersons(req)
}

// GetManagePersons 查询管辖某人员的管理者列表
func (s *OpenPersonService) GetManagePersons(req *model.OpenManagePersonsRequest) (*model.OpenManagePersonsResponse, error) {
	items, err := s.repo.GetManagePersons(req.UniversityID, req.PersonID)
	if err != nil {
		return nil, err
	}
	return &model.OpenManagePersonsResponse{ManagePersons: items}, nil
}

// GetStaffByOrg 按组织机构查询政工（OR条件）
func (s *OpenPersonService) GetStaffByOrg(req *model.OpenStaffByOrgRequest) (*model.OpenStaffResponse, error) {
	return s.repo.GetStaffByOrg(req)
}

// GetStudentByOrg 按组织机构查询学生（OR条件）
func (s *OpenPersonService) GetStudentByOrg(req *model.OpenStudentByOrgRequest) (*model.OpenStudentResponse, error) {
	return s.repo.GetStudentByOrg(req)
}
