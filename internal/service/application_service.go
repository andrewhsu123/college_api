package service

import (
	"college_api/internal/model"
	"college_api/internal/repository"
)

// ApplicationService 应用业务逻辑层
type ApplicationService struct {
	repo *repository.ApplicationRepository
}

// NewApplicationService 创建应用服务
func NewApplicationService(repo *repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

// GetApplicationList 获取应用列表
func (s *ApplicationService) GetApplicationList(req *model.ApplicationListRequest) ([]model.ApplicationWithScope, error) {
	apps, err := s.repo.GetApplicationList(req)
	if err != nil {
		return nil, err
	}

	if len(apps) == 0 {
		return []model.ApplicationWithScope{}, nil
	}

	// 获取应用ID列表
	appIDs := make([]int, len(apps))
	for i, app := range apps {
		appIDs[i] = app.ID
	}

	// 获取应用范围
	scopeMap, err := s.repo.GetApplicationScopes(req.CustomerID, appIDs)
	if err != nil {
		return nil, err
	}

	// 组装结果
	results := make([]model.ApplicationWithScope, len(apps))
	for i, app := range apps {
		results[i] = model.ApplicationWithScope{
			Application: app,
			Scope:       scopeMap[app.ID],
		}
	}

	return results, nil
}

// GetVisibleApplications 获取可见应用列表
func (s *ApplicationService) GetVisibleApplications(req *model.ApplicationVisibleRequest) ([]model.ApplicationWithScope, error) {
	return s.repo.GetVisibleApplications(req)
}
