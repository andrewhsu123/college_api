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

		// 验证Token并获取person_id和customer_id
		personID, customerID, err := m.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": err.Error(),
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将person_id和customer_id存入上下文
		c.Set("person_id", personID)
		c.Set("customer_id", customerID)
		c.Next()
	}
}

// validateToken 验证Token并返回person_id和customer_id
func (m *AuthMiddleware) validateToken(plainToken string) (int64, int, error) {
	// 检查数据库连接
	if err := m.db.Ping(); err != nil {
		fmt.Printf("[STAFF DEBUG] 数据库连接失败: %v\n", err)
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
	fmt.Printf("[STAFF DEBUG] Token ID部分: %s\n", parts[0])
	fmt.Printf("[STAFF DEBUG] Token明文部分: %s\n", parts[1])
	fmt.Printf("[STAFF DEBUG] Token SHA256哈希: %s\n", tokenHash)

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
			fmt.Printf("[STAFF DEBUG] Token在数据库中不存在，哈希值: %s\n", tokenHash)
			return 0, 0, fmt.Errorf("Token无效")
		}
		fmt.Printf("[STAFF DEBUG] 数据库查询错误: %v\n", err)
		return 0, 0, fmt.Errorf("Token验证失败: %v", err)
	}

	fmt.Printf("[STAFF DEBUG] 找到Token记录 - PersonID: %d, Type: %s\n", personID, tokenableType)

	// 验证tokenable_type是否为Person模型
	// 支持多种可能的命名空间格式
	validTypes := []string{
		"App\\Models\\Person",
		"Modules\\Persons\\Models\\Persons",
		"Modules\\Persons\\Models\\Person",
	}

	isValidType := false
	for _, validType := range validTypes {
		if tokenableType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		fmt.Printf("[STAFF DEBUG] Token类型不匹配，期望: %v, 实际: %s\n", validTypes, tokenableType)
		return 0, 0, fmt.Errorf("Token类型错误: 此Token属于%s，不能访问政工端接口", tokenableType)
	}

	// 检查Token是否过期
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		fmt.Printf("[STAFF DEBUG] Token已过期，过期时间: %v\n", expiresAt.Time)
		return 0, 0, fmt.Errorf("Token已过期")
	}

	// 验证person是否存在且为政工类型
	var personType int
	var status int
	var customerID int
	personQuery := `
		SELECT person_type, status, customer_id
		FROM persons
		WHERE id = ? AND deleted_at = 0
		LIMIT 1
	`

	err = m.db.QueryRow(personQuery, personID).Scan(&personType, &status, &customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("[STAFF DEBUG] Person不存在或已删除，PersonID: %d\n", personID)
			return 0, 0, fmt.Errorf("用户不存在")
		}
		fmt.Printf("[STAFF DEBUG] Person查询错误: %v\n", err)
		return 0, 0, fmt.Errorf("用户验证失败: %v", err)
	}

	fmt.Printf("[STAFF DEBUG] Person信息 - Type: %d, Status: %d, CustomerID: %d\n", personType, status, customerID)

	// 验证是否为政工类型 (person_type = 2)
	if personType != 2 {
		fmt.Printf("[STAFF DEBUG] Person类型不匹配，期望: 2(政工), 实际: %d\n", personType)
		return 0, 0, fmt.Errorf("无权限访问: 此账号不是政工类型")
	}

	// 验证用户状态
	if status != 1 {
		fmt.Printf("[STAFF DEBUG] Person状态异常，status: %d\n", status)
		return 0, 0, fmt.Errorf("用户已被禁用")
	}

	// 更新Token最后使用时间
	updateQuery := `UPDATE personal_access_tokens SET last_used_at = ? WHERE token = ?`
	_, _ = m.db.Exec(updateQuery, time.Now(), tokenHash)

	fmt.Printf("[STAFF DEBUG] Token验证成功，PersonID: %d, CustomerID: %d\n", personID, customerID)
	return personID, customerID, nil
}
