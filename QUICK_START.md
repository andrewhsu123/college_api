# College API 快速开始

## 🎯 这是什么？

**Token 验证服务中台** - 只做一件事：验证 Token，返回用户 ID

## ⚡ 5 分钟快速启动

### 1. 配置数据库

```bash
cp .env.example .env
# 编辑 .env，配置数据库连接
```

### 2. 启动服务

```bash
go mod download
go run main.go
```

### 3. 测试

```bash
# 健康检查
curl http://localhost:8080/health

# 验证政工端 Token（需要先从 staff_member_api 获取 Token）
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/staff/info

# 验证学校后台 Token（需要先从 catch_admin_base 获取 Token）
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/base/info
```

## 📝 API 接口

### 政工端验证
```
GET /api/staff/info
Header: Authorization: Bearer {token}

成功响应:
{
  "code": 0,
  "message": "success",
  "data": {
    "person_id": 123
  }
}
```

### 学校后台验证
```
GET /api/base/info
Header: Authorization: Bearer {token}

成功响应:
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 456
  }
}
```

## ❌ 本服务不提供

- 登录接口（去 staff_member_api 或 catch_admin_base）
- 用户注册
- 密码管理
- 用户信息查询

## ✅ 本服务只提供

- Token 验证
- 返回用户 ID
- 就这么简单！

## 📚 更多文档

- [完整 README](README.md)
- [API 测试文档](develop/API_TEST.md)
- [部署文档](develop/DEPLOYMENT.md)

## 🔧 常见问题

**Q: 如何获取 Token？**
A: 在 staff_member_api 或 catch_admin_base 登录后获取

**Q: 为什么只返回 ID？**
A: 这是验证中台，其他服务拿到 ID 后自行查询数据库

**Q: 可以添加登录功能吗？**
A: 不可以，这违背了服务的单一职责原则

**Q: Token 格式是什么？**
A: Laravel Sanctum 格式：`{id}|{plainToken}`

**Q: 如何在其他服务中使用？**
A: 
```go
// 1. 调用本服务验证 Token
resp := callCollegeAPI("/api/staff/info", token)
personID := resp.Data.PersonID

// 2. 查询数据库获取用户信息
var person Person
db.First(&person, personID)

// 3. 处理业务逻辑
// ...
```
