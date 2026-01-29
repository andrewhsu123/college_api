# Go 人员中心认证服务

这是一个与 Laravel CatchAdmin 系统兼容的 Go 语言人员认证加密算法实现。

## 功能特性

### 1. 密码加密（PasswordHasher）

- **完全兼容 Laravel Hash 门面**
- 使用 bcrypt 算法（与 Laravel 一致）
- 支持密码哈希、验证和重新哈希检查

**核心方法：**
- `HashPassword(password string)` - 加密密码（对应 Laravel 的 `Hash::make()`）
- `CheckPassword(password, hashedPassword string)` - 验证密码（对应 Laravel 的 `Hash::check()`）
- `NeedsRehash(hashedPassword string)` - 检查是否需要重新哈希（对应 Laravel 的 `Hash::needsRehash()`）

### 2. Token 生成与解析（TokenGenerator）

- **完全兼容 Laravel PersonsService 的 Token 机制**
- Token 格式：`base64(json_data).md5(signature)`
- 支持过期时间验证
- 支持多角色（student、staff、worker）

**核心方法：**
- `GenerateToken(personID, role, expiresInSeconds)` - 生成 Token
- `ParseToken(token)` - 解析 Token
- `ValidateToken(token)` - 验证 Token 有效性

### 3. 完整认证服务（PersonAuthService）

提供开箱即用的认证服务，包括：
- 用户登录
- Token 验证
- 修改密码
- 用户注册

## 安装依赖

```bash
cd go_code_demo
go mod download
```

## 使用示例

### 基础用法

```go
package main

import (
    "fmt"
    "go_code_demo/auth"
)

func main() {
    // 1. 初始化服务（使用你的 Laravel APP_KEY）
    appKey := "base64:your_laravel_app_key_here"
    authService := auth.NewPersonAuthService(appKey)

    // 2. 用户登录
    loginResp, err := authService.Login("20210001", "123456")
    if err != nil {
        panic(err)
    }
    
    if loginResp.Success {
        fmt.Println("登录成功！")
        data := loginResp.Data.(map[string]interface{})
        token := data["token"].(string)
        fmt.Printf("Token: %s\n", token)
        
        // 3. 验证 Token
        user, err := authService.ValidateToken(token)
        if err != nil {
            fmt.Printf("Token 验证失败: %v\n", err)
        } else {
            fmt.Printf("用户信息: %+v\n", user)
        }
    }
}
```

### 密码加密

```go
hasher := auth.NewPasswordHasher()

// 加密密码
hashedPassword, err := hasher.HashPassword("123456")
if err != nil {
    panic(err)
}
fmt.Println("加密后的密码:", hashedPassword)

// 验证密码
isValid := hasher.CheckPassword("123456", hashedPassword)
fmt.Println("密码验证结果:", isValid)
```

### Token 操作

```go
generator := auth.NewTokenGenerator("your_app_key")

// 生成 Token（7 天有效期）
token, err := generator.GenerateToken(12345, "student", 7*24*3600)
if err != nil {
    panic(err)
}
fmt.Println("生成的 Token:", token)

// 解析 Token
tokenData, err := generator.ParseToken(token)
if err != nil {
    panic(err)
}
fmt.Printf("PersonID: %d, Role: %s\n", tokenData.PersonID, tokenData.Role)

// 验证 Token
personID, role, err := generator.ValidateToken(token)
if err != nil {
    panic(err)
}
fmt.Printf("验证成功 - PersonID: %d, Role: %s\n", personID, role)
```

## 运行测试

```bash
# 运行所有测试
go test ./auth -v

# 运行性能测试
go test ./auth -bench=. -benchmem

# 测试覆盖率
go test ./auth -cover
```

## 与 Laravel 的兼容性

### 密码哈希兼容性

Go 的 bcrypt 实现与 Laravel 的 bcrypt 完全兼容：
- Laravel 使用 `$2y$` 前缀
- Go 使用 `$2a$` 前缀
- 两者可以互相验证

**测试方法：**

1. 在 Laravel 中生成密码哈希：
```php
$hash = Hash::make('password');
echo $hash;
```

2. 在 Go 中验证：
```go
hasher := auth.NewPasswordHasher()
isValid := hasher.CheckPassword("password", "$2y$12$...")
```

### Token 兼容性

Token 格式完全兼容 Laravel PersonsService：
- 相同的 Base64 编码
- 相同的 MD5 签名算法
- 相同的 JSON 数据结构

**注意：** 确保使用相同的 `APP_KEY`

## API 响应格式

所有响应遵循 Laravel 的统一格式：

```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {
    "token": "eyJ...",
    "user": {
      "id": 1,
      "name": "张三",
      "role": "student"
    }
  }
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 10001 | 账号或密码错误 |
| 10002 | 账户已被禁用 |
| 10003 | 学籍状态异常 |
| 10004 | 用户不存在 |
| 10006 | 原密码错误 |
| 10007 | 密码修改失败 |

## 支持的角色类型

- `student` - 学生
- `staff` - 政工
- `worker` - 维修工

## 配置说明

### APP_KEY 获取

从 Laravel 项目的 `.env` 文件中获取：
```
APP_KEY=base64:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### Token 过期时间

默认 7 天（604800 秒），可自定义：
```go
// 自定义过期时间为 24 小时
token, _ := generator.GenerateToken(personID, role, 24*3600)
```

## 性能优化

- bcrypt 成本因子默认为 10（与 Laravel 一致）
- Token 生成和解析使用高效的 JSON 序列化
- 支持并发安全

## 注意事项

1. **APP_KEY 安全**：不要在代码中硬编码 APP_KEY，使用环境变量
2. **密码强度**：建议密码长度至少 8 位，包含字母和数字
3. **Token 存储**：客户端应安全存储 Token（如 HTTPS、加密存储）
4. **过期处理**：Token 过期后需要重新登录

## 集成到中台服务

### HTTP 中间件示例

```go
func AuthMiddleware(authService *auth.PersonAuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 从 Header 获取 Token
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "未授权", http.StatusUnauthorized)
                return
            }
            
            // 验证 Token
            user, err := authService.ValidateToken(token)
            if err != nil {
                http.Error(w, "Token 无效", http.StatusUnauthorized)
                return
            }
            
            // 将用户信息存入 Context
            ctx := context.WithValue(r.Context(), "user", user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## 许可证

与 CatchAdmin 项目保持一致
