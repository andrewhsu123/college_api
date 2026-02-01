package staff

import (
	"college_api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DepartmentHandler 政工端机构处理器
type DepartmentHandler struct {
	service *service.DepartmentService
}

// NewDepartmentHandler 创建政工端机构处理器
func NewDepartmentHandler(service *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

// GetDepartmentTree 获取机构树（仅查看有权限的机构）
func (h *DepartmentHandler) GetDepartmentTree(c *gin.Context) {
	println("[DEBUG] GetDepartmentTree: 开始处理请求")
	
	// 从上下文获取 customer_id 和 person_id
	customerID, exists := c.Get("customer_id")
	if !exists {
		println("[ERROR] GetDepartmentTree: customer_id 不存在")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	personID, exists := c.Get("person_id")
	if !exists {
		println("[ERROR] GetDepartmentTree: person_id 不存在")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	// 类型转换：customer_id 是 int，person_id 是 int64
	custID, ok := customerID.(int)
	if !ok {
		println("[ERROR] GetDepartmentTree: customer_id 类型转换失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "customer_id 类型错误",
			"data":    nil,
		})
		return
	}
	
	persID64, ok := personID.(int64)
	if !ok {
		println("[ERROR] GetDepartmentTree: person_id 类型转换失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "person_id 类型错误",
			"data":    nil,
		})
		return
	}
	persID := int(persID64)
	
	println("[DEBUG] GetDepartmentTree: customer_id =", custID, ", person_id =", persID)

	// 获取政工人员可见的机构ID列表
	println("[DEBUG] GetDepartmentTree: 开始获取可见机构ID列表")
	visibleDeptIDs, err := h.service.GetStaffVisibleDepartmentIDs(custID, persID)
	if err != nil {
		println("[ERROR] GetDepartmentTree: 获取权限失败:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限失败",
			"error":   err.Error(),
		})
		return
	}
	println("[DEBUG] GetDepartmentTree: 可见机构ID数量 =", len(visibleDeptIDs))

	// 查询机构树（带权限过滤）
	println("[DEBUG] GetDepartmentTree: 开始查询机构树")
	tree, err := h.service.GetDepartmentTree(custID, visibleDeptIDs)
	if err != nil {
		println("[ERROR] GetDepartmentTree: 查询失败:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
			"error":   err.Error(),
		})
		return
	}
	println("[DEBUG] GetDepartmentTree: 查询成功，返回树节点数量 =", len(tree))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    tree,
	})
	println("[DEBUG] GetDepartmentTree: 请求处理完成")
}

// GetDepartmentList 获取机构列表（仅查看有权限的机构）
func (h *DepartmentHandler) GetDepartmentList(c *gin.Context) {
	// 从上下文获取 customer_id 和 person_id
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	personID, exists := c.Get("person_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	// 类型转换：customer_id 是 int，person_id 是 int64
	custID, ok := customerID.(int)
	if !ok {
		println("[ERROR] GetDepartmentList: customer_id 类型转换失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "customer_id 类型错误",
			"data":    nil,
		})
		return
	}
	
	persID64, ok := personID.(int64)
	if !ok {
		println("[ERROR] GetDepartmentList: person_id 类型转换失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "person_id 类型错误",
			"data":    nil,
		})
		return
	}
	persID := int(persID64)
	
	// 获取查询参数
	keyword := c.Query("keyword")
	
	var departmentType *int
	if typeStr := c.Query("department_type"); typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil {
			departmentType = &t
		}
	}

	// 获取政工人员可见的机构ID列表
	visibleDeptIDs, err := h.service.GetStaffVisibleDepartmentIDs(custID, persID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限失败",
			"error":   err.Error(),
		})
		return
	}

	// 查询机构列表（带权限过滤）
	list, err := h.service.GetDepartmentList(custID, keyword, departmentType, visibleDeptIDs)
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
