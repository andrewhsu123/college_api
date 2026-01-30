package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler 学校后台认证处理器
type AuthHandler struct{}

// NewAuthHandler 创建学校后台认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// GetUserInfo 获取当前登录用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// 从上下文获取user_id
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user_id": userID,
		},
	})
}
