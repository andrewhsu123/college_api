package base

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

// AuthMiddleware 学校后台Token验证中间件
type AuthMiddleware struct {
	db *sql.DB
}

// NewAuthMiddleware 创建学校后台认证中间件
func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// Authenticate 验证Token并返回user_id
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

		// 验证Token并获取user_id和customer_id
		userID, customerID, err := m.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将user_id和customer_id存入上下文
		c.Set("user_id", userID)
		c.Set("customer_id", customerID)
		c.Next()
	}
}

// validateToken 验证Token并返回user_id和customer_id
func (m *AuthMiddleware) validateToken(plainToken string) (int64, int, error) {
	// 检查数据库连接
	if err := m.db.Ping(); err != nil {
		fmt.Printf("[DEBUG] 数据库连接失败: %v\n", err)
		return 0, 0, fmt.Errorf("数据库连接失败，请稍后重试")
	}

	// Laravel Sanctum Token格式: {id}|{plainToken}
	// 实际存储的是SHA256哈希值
	parts := strings.SplitN(plainToken, "|", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("Token格式错误: 应为 {id}|{token} 格式")
	}

	// 计算Token的SHA256哈希
	hash := sha256.Sum256([]byte(parts[1]))
	tokenHash := hex.EncodeToString(hash[:])

	// 调试信息
	fmt.Printf("[DEBUG] Token ID部分: %s\n", parts[0])
	fmt.Printf("[DEBUG] Token明文部分: %s\n", parts[1])
	fmt.Printf("[DEBUG] Token SHA256哈希: %s\n", tokenHash)

	// 查询数据库验证Token
	var userID int64
	var expiresAt sql.NullTime
	var tokenableType string

	query := `
		SELECT tokenable_id, expires_at, tokenable_type
		FROM personal_access_tokens
		WHERE token = ?
		LIMIT 1
	`

	err := m.db.QueryRow(query, tokenHash).Scan(&userID, &expiresAt, &tokenableType)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("[DEBUG] Token在数据库中不存在，哈希值: %s\n", tokenHash)
			// 尝试查询最近的几条记录对比
			var count int
			m.db.QueryRow("SELECT COUNT(*) FROM personal_access_tokens").Scan(&count)
			fmt.Printf("[DEBUG] 数据库中共有 %d 条Token记录\n", count)
			return 0, 0, fmt.Errorf("Token无效")
		}
		fmt.Printf("[DEBUG] 数据库查询错误: %v\n", err)
		return 0, 0, fmt.Errorf("Token验证失败: %v", err)
	}

	fmt.Printf("[DEBUG] 找到Token记录 - UserID: %d, Type: %s\n", userID, tokenableType)

	// 验证tokenable_type是否为User模型
	if tokenableType != "Modules\\User\\Models\\User" {
		fmt.Printf("[DEBUG] Token类型不匹配，期望: Modules\\User\\Models\\User, 实际: %s\n", tokenableType)
		return 0, 0, fmt.Errorf("Token类型错误")
	}

	// 检查Token是否过期
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		fmt.Printf("[DEBUG] Token已过期，过期时间: %v\n", expiresAt.Time)
		return 0, 0, fmt.Errorf("Token已过期")
	}

	// 验证用户是否存在
	var status int
	var customerID int
	userQuery := `
		SELECT status, customer_id
		FROM admin_users
		WHERE id = ? AND deleted_at = 0
		LIMIT 1
	`

	err = m.db.QueryRow(userQuery, userID).Scan(&status, &customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("[DEBUG] 用户不存在或已删除，UserID: %d\n", userID)
			return 0, 0, fmt.Errorf("用户不存在")
		}
		fmt.Printf("[DEBUG] 用户查询错误: %v\n", err)
		return 0, 0, fmt.Errorf("用户验证失败: %v", err)
	}

	fmt.Printf("[DEBUG] 用户状态: %d, CustomerID: %d\n", status, customerID)

	// 验证用户状态 (status = 1 正常)
	if status != 1 {
		fmt.Printf("[DEBUG] 用户状态异常，status: %d\n", status)
		return 0, 0, fmt.Errorf("用户已被禁用")
	}

	// 更新Token最后使用时间
	updateQuery := `UPDATE personal_access_tokens SET last_used_at = ? WHERE token = ?`
	_, _ = m.db.Exec(updateQuery, time.Now(), tokenHash)

	fmt.Printf("[DEBUG] Token验证成功，UserID: %d, CustomerID: %d\n", userID, customerID)
	
	return userID, customerID, nil
}
