package base

import (
	"college_api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DepartmentHandler 学校后台机构处理器
type DepartmentHandler struct {
	service *service.DepartmentService
}

// NewDepartmentHandler 创建学校后台机构处理器
func NewDepartmentHandler(service *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

// GetDepartmentTree 获取机构树
func (h *DepartmentHandler) GetDepartmentTree(c *gin.Context) {
	// 从上下文获取 customer_id
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	// 学校管理员查看所有机构（不传 visibleDeptIDs）
	tree, err := h.service.GetDepartmentTree(customerID.(int), nil)
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
		"data":    tree,
	})
}

// GetDepartmentList 获取机构列表
func (h *DepartmentHandler) GetDepartmentList(c *gin.Context) {
	// 从上下文获取 customer_id
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	// 获取查询参数
	keyword := c.Query("keyword")
	
	var departmentType *int
	if typeStr := c.Query("department_type"); typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil {
			departmentType = &t
		}
	}

	// 学校管理员查看所有机构（不传 visibleDeptIDs）
	list, err := h.service.GetDepartmentList(customerID.(int), keyword, departmentType, nil)
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
		"data":    list,
	})
}
