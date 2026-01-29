package auth

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher 密码哈希处理器
type PasswordHasher struct {
	cost int // bcrypt 成本因子，默认 10
}

// NewPasswordHasher 创建密码哈希处理器
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost: bcrypt.DefaultCost, // 默认 10，与 Laravel 一致
	}
}

// HashPassword 对密码进行哈希加密（使用 bcrypt，与 Laravel Hash::make 兼容）
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码是否匹配（与 Laravel Hash::check 兼容）
func (h *PasswordHasher) CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// NeedsRehash 检查密码是否需要重新哈希
func (h *PasswordHasher) NeedsRehash(hashedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		return true
	}
	return cost != h.cost
}

// TokenData Token 数据结构
type TokenData struct {
	PersonID  int64  `json:"person_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

// TokenGenerator Token 生成器
type TokenGenerator struct {
	appKey string // 应用密钥，对应 Laravel 的 APP_KEY
}

// NewTokenGenerator 创建 Token 生成器
func NewTokenGenerator(appKey string) *TokenGenerator {
	return &TokenGenerator{
		appKey: appKey,
	}
}

// GenerateToken 生成 Token（与 Laravel PersonsService::generateToken 兼容）
// 格式: base64(json_data).md5(person_id + role + timestamp + app_key)
func (g *TokenGenerator) GenerateToken(personID int64, role string, expiresInSeconds int64) (string, error) {
	if expiresInSeconds <= 0 {
		expiresInSeconds = 7 * 24 * 3600 // 默认 7 天
	}

	tokenData := TokenData{
		PersonID:  personID,
		Role:      role,
		ExpiresAt: time.Now().Unix() + expiresInSeconds,
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("序列化 token 数据失败: %w", err)
	}

	// Base64 编码
	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	// 生成签名：md5(person_id + role + timestamp + app_key)
	timestamp := time.Now().Unix()
	signatureInput := fmt.Sprintf("%d%s%d%s", personID, role, timestamp, g.appKey)
	signature := md5Hash(signatureInput)

	// 拼接 token
	token := fmt.Sprintf("%s.%s", encodedData, signature)

	return token, nil
}

// ParseToken 解析 Token（与 Laravel PersonsService::parseToken 兼容）
func (g *TokenGenerator) ParseToken(token string) (*TokenData, error) {
	if token == "" {
		return nil, errors.New("token 为空")
	}

	// 分割 token
	var encodedData, signature string
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			encodedData = token[:i]
			signature = token[i+1:]
			break
		}
	}

	if encodedData == "" || signature == "" {
		return nil, errors.New("token 格式错误")
	}

	// Base64 解码
	jsonData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("token 解码失败: %w", err)
	}

	// 反序列化 JSON
	var tokenData TokenData
	if err := json.Unmarshal(jsonData, &tokenData); err != nil {
		return nil, fmt.Errorf("token 数据解析失败: %w", err)
	}

	// 验证必要字段
	if tokenData.PersonID == 0 || tokenData.Role == "" {
		return nil, errors.New("token 数据不完整")
	}

	// 检查过期时间
	if tokenData.ExpiresAt > 0 && tokenData.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token 已过期")
	}

	return &tokenData, nil
}

// ValidateToken 验证 Token 是否有效
func (g *TokenGenerator) ValidateToken(token string) (int64, string, error) {
	tokenData, err := g.ParseToken(token)
	if err != nil {
		return 0, "", err
	}
	return tokenData.PersonID, tokenData.Role, nil
}

// md5Hash 计算 MD5 哈希
func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
