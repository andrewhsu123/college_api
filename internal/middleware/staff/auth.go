package staff

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 政工端Token验证中间件
type AuthMiddleware struct {
	db *sql.DB
}

// NewAuthMiddleware 创建政工端认证中间件
func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// Authenticate 验证Token并返回person_id
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 解析Token (格式: Bearer {token})
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token格式错误",
				"data":    nil,
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证Token并获取person_id
		personID, err := m.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将person_id存入上下文
		c.Set("person_id", personID)
		c.Next()
	}
}

// validateToken 验证Token并返回person_id
func (m *AuthMiddleware) validateToken(plainToken string) (int64, error) {
	// Laravel Sanctum Token格式: {id}|{plainToken}
	// 实际存储的是SHA256哈希值
	parts := strings.SplitN(plainToken, "|", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("Token格式错误")
	}

	// 计算Token的SHA256哈希
	hash := sha256.Sum256([]byte(parts[1]))
	tokenHash := hex.EncodeToString(hash[:])

	// 查询数据库验证Token
	var personID int64
	var expiresAt sql.NullTime
	var tokenableType string

	query := `
		SELECT tokenable_id, expires_at, tokenable_type
		FROM personal_access_tokens
		WHERE token = ?
		LIMIT 1
	`

	err := m.db.QueryRow(query, tokenHash).Scan(&personID, &expiresAt, &tokenableType)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("Token无效")
		}
		return 0, fmt.Errorf("Token验证失败: %v", err)
	}

	// 验证tokenable_type是否为Person模型
	if tokenableType != "App\\Models\\Person" {
		return 0, fmt.Errorf("Token类型错误")
	}

	// 检查Token是否过期
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return 0, fmt.Errorf("Token已过期")
	}

	// 验证person是否存在且为政工类型
	var personType int
	var status int
	personQuery := `
		SELECT person_type, status
		FROM persons
		WHERE id = ? AND deleted_at = 0
		LIMIT 1
	`

	err = m.db.QueryRow(personQuery, personID).Scan(&personType, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("用户不存在")
		}
		return 0, fmt.Errorf("用户验证失败: %v", err)
	}

	// 验证是否为政工类型 (person_type = 2)
	if personType != 2 {
		return 0, fmt.Errorf("无权限访问")
	}

	// 验证用户状态
	if status != 1 {
		return 0, fmt.Errorf("用户已被禁用")
	}

	// 更新Token最后使用时间
	updateQuery := `UPDATE personal_access_tokens SET last_used_at = ? WHERE token = ?`
	_, _ = m.db.Exec(updateQuery, time.Now(), tokenHash)

	return personID, nil
}
