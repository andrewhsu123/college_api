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
