package staff

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler 政工端认证处理器
type AuthHandler struct{}

// NewAuthHandler 创建政工端认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// GetPersonInfo 获取当前登录政工信息
func (h *AuthHandler) GetPersonInfo(c *gin.Context) {
	// 从上下文获取person_id
	personID, exists := c.Get("person_id")
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
			"person_id": personID,
		},
	})
}
