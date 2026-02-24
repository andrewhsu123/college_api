package open

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 开放接口认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从环境变量获取秘钥
		openKey := os.Getenv("OPEN_KEY")
		if openKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "服务配置错误",
			})
			c.Abort()
			return
		}

		// 从 Authorization 头获取秘钥
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证信息",
			})
			c.Abort()
			return
		}

		// 支持 Bearer token 格式
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 验证秘钥
		if token != openKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "秘钥无效",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
