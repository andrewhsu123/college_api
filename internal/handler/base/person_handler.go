package base

import (
	"college_api/internal/model"
	"college_api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PersonHandler 学校管理员人员处理器
type PersonHandler struct {
	service *service.PersonService
}

// NewPersonHandler 创建人员处理器
func NewPersonHandler(service *service.PersonService) *PersonHandler {
	return &PersonHandler{service: service}
}

// GetPersonList 查询人员列表
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

	// 学校管理员可以查看所有人员，不需要权限过滤
	resp, err := h.service.GetPersonList(&req, false, nil, nil)
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
