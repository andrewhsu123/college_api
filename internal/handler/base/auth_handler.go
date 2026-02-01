package base

import (
	"college_api/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler 学校后台认证处理器
type AuthHandler struct {
	personService *service.PersonService
}

// NewAuthHandler 创建学校后台认证处理器
func NewAuthHandler(personService *service.PersonService) *AuthHandler {
	return &AuthHandler{
		personService: personService,
	}
}

// GetUserInfo 获取当前登录用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// 从上下文获取user_id
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"data":    nil,
		})
		return
	}

	// 调试：打印类型
	fmt.Printf("[DEBUG] GetUserInfo: user_id type = %T, value = %v\n", userIDInterface, userIDInterface)

	// 类型转换：middleware 返回的是 int64
	var userID int
	switch v := userIDInterface.(type) {
	case int64:
		userID = int(v)
		fmt.Printf("[DEBUG] GetUserInfo: 从 int64 转换, userID = %d\n", userID)
	case int:
		userID = v
		fmt.Printf("[DEBUG] GetUserInfo: 直接使用 int, userID = %d\n", userID)
	default:
		fmt.Printf("[ERROR] GetUserInfo: user_id 类型转换失败, type = %T\n", userIDInterface)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "user_id 类型错误",
			"data":    nil,
		})
		return
	}

	// 查询管理员完整信息
	info, err := h.personService.GetAdminUserInfo(userID)
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
