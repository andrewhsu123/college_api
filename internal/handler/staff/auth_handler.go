package staff

import (
	"college_api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler 政工端认证处理器
type AuthHandler struct {
	personService *service.PersonService
}

// NewAuthHandler 创建政工端认证处理器
func NewAuthHandler(personService *service.PersonService) *AuthHandler {
	return &AuthHandler{
		personService: personService,
	}
}

// GetPersonInfo 获取当前登录政工信息
func (h *AuthHandler) GetPersonInfo(c *gin.Context) {
	// 从上下文获取person_id
	personIDInterface, exists := c.Get("person_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"data":    nil,
		})
		return
	}

	// 类型转换：person_id 是 int64
	personID64, ok := personIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "person_id 类型错误",
			"data":    nil,
		})
		return
	}
	personID := int(personID64)

	// 查询政工完整信息
	info, err := h.personService.GetStaffInfo(personID)
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
		"data":    info,
	})
}
