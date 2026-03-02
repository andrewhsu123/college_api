package service

import (
	"college_api/internal/model"
	"college_api/internal/repository"
)

// CollegeService 学校服务
type CollegeService struct {
	repo *repository.CollegeRepository
}

// NewCollegeService 创建学校服务
func NewCollegeService(repo *repository.CollegeRepository) *CollegeService {
	return &CollegeService{repo: repo}
}

// GetCollegeList 查询学校列表
func (s *CollegeService) GetCollegeList(req *model.OpenCollegeRequest) (*model.OpenCollegeResponse, error) {
	return s.repo.GetCollegeList(req)
}

// GetCampusAreaList 查询校区列表
func (s *CollegeService) GetCampusAreaList(req *model.OpenCampusAreaRequest) ([]model.OpenCampusAreaItem, error) {
	return s.repo.GetCampusAreaList(req)
}

// GetDepartmentList 查询部门列表
func (s *CollegeService) GetDepartmentList(req *model.OpenDepartmentListRequest) ([]model.OpenDepartmentListItem, error) {
	return s.repo.GetDepartmentList(req)
}
