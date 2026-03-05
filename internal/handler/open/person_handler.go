package open

import (
	"college_api/internal/model"
	"college_api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PersonHandler 开放接口人员处理器
type PersonHandler struct {
	service *service.OpenPersonService
}

// NewPersonHandler 创建人员处理器
func NewPersonHandler(service *service.OpenPersonService) *PersonHandler {
	return &PersonHandler{service: service}
}

// GetStaffList 查询政工列表
func (h *PersonHandler) GetStaffList(c *gin.Context) {
	var req model.OpenStaffRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.service.GetStaffList(&req)
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

// GetStudentList 查询学生列表
func (h *PersonHandler) GetStudentList(c *gin.Context) {
	var req model.OpenStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.service.GetStudentList(&req)
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

// GetRoleList 查询角色列表
func (h *PersonHandler) GetRoleList(c *gin.Context) {
	var req model.OpenRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	items, err := h.service.GetRoleList(&req)
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

// GetManagePersons 查询管辖某人员的管理者列表
func (h *PersonHandler) GetManagePersons(c *gin.Context) {
	var req model.OpenManagePersonsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.service.GetManagePersons(&req)
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

// GetRolePersons 根据角色查询人员详情列表
func (h *PersonHandler) GetRolePersons(c *gin.Context) {
	var req model.OpenRolePersonsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if req.RoleIDs == "" && req.RoleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   "role_ids or role_name is required",
		})
		return
	}

	resp, err := h.service.GetRolePersons(&req)
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

// GetStaffByOrg 按组织机构查询政工（OR条件）
// GET /open/staff/by-org?university_id=1&department_ids=1,2,3&college_ids=4,5&faculty_ids=6
func (h *PersonHandler) GetStaffByOrg(c *gin.Context) {
	var req model.OpenStaffByOrgRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 至少需要一个组织条件
	if req.DepartmentIDs == "" && req.CollegeIDs == "" && req.FacultyIDs == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   "at least one of department_ids, college_ids, faculty_ids is required",
		})
		return
	}

	resp, err := h.service.GetStaffByOrg(&req)
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

// GetStudentByOrg 按组织机构查询学生（OR条件）
// GET /open/student/by-org?university_id=1&college_ids=1,2&faculty_ids=3&profession_ids=4,5&class_ids=6,7
func (h *PersonHandler) GetStudentByOrg(c *gin.Context) {
	var req model.OpenStudentByOrgRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 至少需要一个组织条件
	if req.CollegeIDs == "" && req.FacultyIDs == "" && req.ProfessionIDs == "" && req.ClassIDs == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   "at least one of college_ids, faculty_ids, profession_ids, class_ids is required",
		})
		return
	}

	resp, err := h.service.GetStudentByOrg(&req)
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
