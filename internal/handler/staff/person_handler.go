package staff

import (
	"college_api/internal/model"
	"college_api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PersonHandler 政工人员处理器
type PersonHandler struct {
	service *service.PersonService
}

// NewPersonHandler 创建人员处理器
func NewPersonHandler(service *service.PersonService) *PersonHandler {
	return &PersonHandler{service: service}
}

// GetPersonList 查询人员列表（带权限过滤）
func (h *PersonHandler) GetPersonList(c *gin.Context) {
	var req model.PersonListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1001,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 从上下文获取政工人员信息
	staffInfo, exists := c.Get("staff_info")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    1002,
			"message": "未授权",
		})
		return
	}

	info := staffInfo.(*model.StaffInfo)

	// 使用缓存的权限信息进行过滤
	resp, err := h.service.GetPersonList(&req, true, info.ManagedDepartmentIDs, info.ManagedPersonIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1005,
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
