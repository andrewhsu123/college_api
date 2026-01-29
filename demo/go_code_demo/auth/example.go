package auth

import (
	"fmt"
	"log"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Account  string `json:"account"`
	Role     string `json:"role"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Status   int    `json:"status"`
}

// PersonAuthService 人员认证服务
type PersonAuthService struct {
	hasher    *PasswordHasher
	generator *TokenGenerator
}

// NewPersonAuthService 创建人员认证服务
func NewPersonAuthService(appKey string) *PersonAuthService {
	return &PersonAuthService{
		hasher:    NewPasswordHasher(),
		generator: NewTokenGenerator(appKey),
	}
}

// Login 用户登录示例
func (s *PersonAuthService) Login(account, password string) (*LoginResponse, error) {
	// 1. 从数据库查询用户（这里用模拟数据）
	user, hashedPassword := s.getUserFromDB(account)
	if user == nil {
		return &LoginResponse{
			Success: false,
			Code:    10001,
			Message: "账号或密码错误",
		}, nil
	}

	// 2. 验证密码
	if !s.hasher.CheckPassword(password, hashedPassword) {
		return &LoginResponse{
			Success: false,
			Code:    10001,
			Message: "账号或密码错误",
		}, nil
	}

	// 3. 检查用户状态
	if user.Status != 1 {
		return &LoginResponse{
			Success: false,
			Code:    10002,
			Message: "该账户已被禁用",
		}, nil
	}

	// 4. 生成 token
	token, err := s.generator.GenerateToken(user.ID, user.Role, 7*24*3600)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	// 5. 返回登录成功响应
	return &LoginResponse{
		Success: true,
		Message: "登录成功",
		Data: map[string]interface{}{
			"token": token,
			"user":  user,
		},
	}, nil
}

// ValidateToken 验证 token
func (s *PersonAuthService) ValidateToken(token string) (*UserInfo, error) {
	// 1. 解析 token
	tokenData, err := s.generator.ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("token 无效: %w", err)
	}

	// 2. 从数据库获取用户信息（这里用模拟数据）
	user := s.getUserByID(tokenData.PersonID)
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 3. 检查用户状态
	if user.Status != 1 {
		return nil, fmt.Errorf("账户已被禁用")
	}

	return user, nil
}

// ChangePassword 修改密码
func (s *PersonAuthService) ChangePassword(personID int64, oldPassword, newPassword string) (*LoginResponse, error) {
	// 1. 获取用户信息
	user, hashedPassword := s.getUserByIDWithPassword(personID)
	if user == nil {
		return &LoginResponse{
			Success: false,
			Code:    10004,
			Message: "用户不存在",
		}, nil
	}

	// 2. 验证原密码
	if !s.hasher.CheckPassword(oldPassword, hashedPassword) {
		return &LoginResponse{
			Success: false,
			Code:    10006,
			Message: "原密码错误",
		}, nil
	}

	// 3. 加密新密码
	newHashedPassword, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新数据库（这里仅示例）
	if err := s.updatePasswordInDB(personID, newHashedPassword); err != nil {
		return &LoginResponse{
			Success: false,
			Code:    10007,
			Message: "密码修改失败，请重试",
		}, nil
	}

	return &LoginResponse{
		Success: true,
		Message: "密码修改成功",
	}, nil
}

// RegisterUser 注册用户（创建用户时加密密码）
func (s *PersonAuthService) RegisterUser(account, password, name, role string) (*UserInfo, error) {
	// 1. 加密密码
	hashedPassword, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 2. 保存到数据库（这里仅示例）
	user := &UserInfo{
		ID:      12345, // 实际应该是数据库生成的 ID
		Name:    name,
		Account: account,
		Role:    role,
		Status:  1,
	}

	// 实际应该保存 hashedPassword 到数据库
	_ = hashedPassword

	return user, nil
}

// 以下是模拟数据库操作的方法

func (s *PersonAuthService) getUserFromDB(account string) (*UserInfo, string) {
	// 模拟数据库查询
	// 实际应该从数据库查询
	if account == "20210001" {
		return &UserInfo{
			ID:      1,
			Name:    "张三",
			Account: "20210001",
			Role:    "student",
			Mobile:  "13800138000",
			Email:   "zhangsan@example.com",
			Status:  1,
		}, "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy" // password: "123456"
	}
	return nil, ""
}

func (s *PersonAuthService) getUserByID(personID int64) *UserInfo {
	// 模拟数据库查询
	if personID == 1 {
		return &UserInfo{
			ID:      1,
			Name:    "张三",
			Account: "20210001",
			Role:    "student",
			Status:  1,
		}
	}
	return nil
}

func (s *PersonAuthService) getUserByIDWithPassword(personID int64) (*UserInfo, string) {
	// 模拟数据库查询
	if personID == 1 {
		return &UserInfo{
			ID:      1,
			Name:    "张三",
			Account: "20210001",
			Role:    "student",
			Status:  1,
		}, "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}
	return nil, ""
}

func (s *PersonAuthService) updatePasswordInDB(personID int64, hashedPassword string) error {
	// 模拟数据库更新
	log.Printf("更新用户 %d 的密码: %s", personID, hashedPassword)
	return nil
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// 初始化服务（使用你的 APP_KEY）
	appKey := "base64:your_laravel_app_key_here"
	authService := NewPersonAuthService(appKey)

	// 示例 1: 用户登录
	fmt.Println("=== 示例 1: 用户登录 ===")
	loginResp, err := authService.Login("20210001", "123456")
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	fmt.Printf("登录响应: %+v\n\n", loginResp)

	// 示例 2: 验证 token
	if loginResp.Success {
		fmt.Println("=== 示例 2: 验证 Token ===")
		data := loginResp.Data.(map[string]interface{})
		token := data["token"].(string)
		fmt.Printf("Token: %s\n", token)

		user, err := authService.ValidateToken(token)
		if err != nil {
			log.Printf("Token 验证失败: %v", err)
		} else {
			fmt.Printf("验证成功，用户信息: %+v\n\n", user)
		}
	}

	// 示例 3: 修改密码
	fmt.Println("=== 示例 3: 修改密码 ===")
	changeResp, err := authService.ChangePassword(1, "123456", "newpassword123")
	if err != nil {
		log.Fatalf("修改密码失败: %v", err)
	}
	fmt.Printf("修改密码响应: %+v\n\n", changeResp)

	// 示例 4: 注册用户
	fmt.Println("=== 示例 4: 注册用户 ===")
	newUser, err := authService.RegisterUser("20210002", "password123", "李四", "student")
	if err != nil {
		log.Fatalf("注册失败: %v", err)
	}
	fmt.Printf("注册成功，用户信息: %+v\n", newUser)
}
