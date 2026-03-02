package open

import (
	"college_api/internal/model"
	"college_api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CollegeHandler 开放接口学校处理器
type CollegeHandler struct {
	service *service.CollegeService
}

// NewCollegeHandler 创建学校处理器
func NewCollegeHandler(service *service.CollegeService) *CollegeHandler {
	return &CollegeHandler{service: service}
}

// GetCollegeList 查询学校列表
func (h *CollegeHandler) GetCollegeList(c *gin.Context) {
	var req model.OpenCollegeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.service.GetCollegeList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// GetCampusAreaList 查询校区列表
func (h *CollegeHandler) GetCampusAreaList(c *gin.Context) {
	var req model.OpenCampusAreaRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	items, err := h.service.GetCampusAreaList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    items,
	})
}

// GetDepartmentList 查询部门列表
func (h *CollegeHandler) GetDepartmentList(c *gin.Context) {
	var req model.OpenDepartmentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	items, err := h.service.GetDepartmentList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    items,
	})
}
