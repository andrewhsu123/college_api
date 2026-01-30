# College API 测试文档

## 服务说明

**College API 是一个纯粹的 Token 验证服务中台**

- ✅ 只验证 Token 有效性
- ✅ 只返回用户身份 ID
- ❌ 不提供登录接口
- ❌ 不提供用户管理接口
- ❌ 不返回用户详细信息

## 测试环境准备

### 1. 启动服务

```bash
cd college_api
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 2. 获取测试 Token

**⚠️ 重要：本服务不提供登录接口！**

Token 需要从以下业务系统获取：

#### 政工端 Token
- 在 `staff_member_api` 项目中登录获取
- 登录接口：`POST /app/login`
- Token 格式示例：`1|abcdefghijklmnopqrstuvwxyz1234567890`

#### 学校后台 Token
- 在 `catch_admin_base` 项目中登录获取
- 登录接口：`POST /api/login`
- Token 格式示例：`2|zyxwvutsrqponmlkjihgfedcba0987654321`

**测试前准备：**
1. 确保 `staff_member_api` 或 `catch_admin_base` 服务正常运行
2. 通过它们的登录接口获取有效 Token
3. 使用获取到的 Token 测试本服务

## API 测试用例

### 1. 健康检查

```bash
curl http://localhost:8080/health
```

**预期响应:**
```json
{
  "status": "ok",
  "message": "College API Service"
}
```

---

### 2. 政工端 - 验证 Token 并获取身份

**功能：验证 Token 是否有效，返回 person_id**

#### 请求示例

```bash
curl -X GET http://localhost:8080/api/staff/info \
  -H "Authorization: Bearer 1|abcdefghijklmnopqrstuvwxyz1234567890"
```

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "person_id": 123
  }
}
```

**说明：**
- 只返回 `person_id`，不返回姓名、手机号等其他信息
- 调用方需要根据 `person_id` 自行查询 `persons` 表获取完整信息

#### 错误响应示例

**未提供 Token (401):**
```json
{
  "code": 401,
  "message": "请先登录",
  "data": null
}
```

**Token 格式错误 (401):**
```json
{
  "code": 401,
  "message": "Token格式错误",
  "data": null
}
```

**Token 无效 (401):**
```json
{
  "code": 401,
  "message": "Token无效",
  "data": null
}
```
**说明：** Token 在数据库中不存在或已被删除

**非政工用户 (401):**
```json
{
  "code": 401,
  "message": "无权限访问",
  "data": null
}
```
**说明：** 该用户不是政工类型 (person_type ≠ 2)

**用户已禁用 (401):**
```json
{
  "code": 401,
  "message": "用户已被禁用",
  "data": null
}
```
**说明：** 用户状态异常 (status ≠ 1)

**Token 已过期 (401):**
```json
{
  "code": 401,
  "message": "Token已过期",
  "data": null
}
```
**说明：** Token 的 expires_at 时间已过

---

### 3. 学校后台 - 验证 Token 并获取身份

**功能：验证 Token 是否有效，返回 user_id**

#### 请求示例

```bash
curl -X GET http://localhost:8080/api/base/info \
  -H "Authorization: Bearer 2|zyxwvutsrqponmlkjihgfedcba0987654321"
```

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 456
  }
}
```

**说明：**
- 只返回 `user_id`，不返回用户名、邮箱等其他信息
- 调用方需要根据 `user_id` 自行查询 `admin_users` 表获取完整信息

#### 错误响应

与政工端类似，但不会检查 person_type。

---

## 使用场景示例

### 场景 1: 微服务调用流程

```
1. 用户在 staff_member_api 登录（不在本服务）
   → 获取 Token: "1|abc123..."

2. 用户访问其他微服务（如订单服务）
   GET /api/orders
   Header: Authorization: Bearer 1|abc123...

3. 订单服务调用 College API 验证 Token
   GET http://college-api:8080/api/staff/info
   Header: Authorization: Bearer 1|abc123...
   → 返回: {"person_id": 123}

4. 订单服务根据 person_id 查询用户信息
   SELECT * FROM persons WHERE id = 123
   → 获取完整用户信息（姓名、手机号等）

5. 订单服务处理业务逻辑并返回结果
```

### 场景 2: 网关集成

```
┌──────────┐     ┌──────────┐     ┌──────────────┐     ┌──────────┐
│  客户端   │────▶│  API网关  │────▶│ College API  │────▶│  数据库   │
└──────────┘     └──────────┘     └──────────────┘     └──────────┘
     │                │                    │                   │
     │  1. 请求       │                    │                   │
     │  + Token       │                    │                   │
     │                │  2. 验证 Token     │                   │
     │                │                    │  3. 查询 Token    │
     │                │                    │                   │
     │                │  4. 返回 user_id   │                   │
     │                │                    │                   │
     │                │  5. 转发请求       │                   │
     │                │  + user_id         │                   │
     │                │                    │                   │
     │  6. 返回结果   │                    │                   │
     └────────────────┘                    │                   │
```

### 场景 3: 权限验证示例代码

```go
// 其他微服务中的示例代码

// 1. 验证 Token 并获取用户 ID
func AuthMiddleware(c *gin.Context) {
    token := c.GetHeader("Authorization")
    
    // 调用 College API 验证
    resp := callCollegeAPI("/api/staff/info", token)
    if resp.Code != 0 {
        c.JSON(401, gin.H{"error": "未授权"})
        c.Abort()
        return
    }
    
    personID := resp.Data.PersonID
    c.Set("person_id", personID)
    c.Next()
}

// 2. 在业务逻辑中使用
func GetOrders(c *gin.Context) {
    personID := c.GetInt64("person_id")
    
    // 查询用户信息
    var person Person
    db.First(&person, personID)
    
    // 查询订单
    var orders []Order
    db.Where("person_id = ?", personID).Find(&orders)
    
    c.JSON(200, orders)
}
```

---

## Postman 测试集合

### 环境变量设置

```json
{
  "base_url": "http://localhost:8080",
  "staff_token": "1|your_staff_token_here",
  "admin_token": "2|your_admin_token_here"
}
```

### 测试集合

#### 1. Health Check
- **Method**: GET
- **URL**: `{{base_url}}/health`
- **Headers**: 无

#### 2. Staff Info
- **Method**: GET
- **URL**: `{{base_url}}/api/staff/info`
- **Headers**: 
  - `Authorization`: `Bearer {{staff_token}}`

#### 3. Admin Info
- **Method**: GET
- **URL**: `{{base_url}}/api/base/info`
- **Headers**: 
  - `Authorization`: `Bearer {{admin_token}}`

---

## 数据库验证

### 查看 Token 记录

```sql
-- 查看所有 Token
SELECT 
    id,
    tokenable_type,
    tokenable_id,
    name,
    LEFT(token, 10) as token_prefix,
    last_used_at,
    expires_at,
    created_at
FROM personal_access_tokens
ORDER BY created_at DESC
LIMIT 10;
```

### 查看政工人员

```sql
-- 查看政工类型的人员
SELECT 
    id,
    name,
    person_type,
    mobile,
    status,
    created_at
FROM persons
WHERE person_type = 2 
  AND deleted_at = 0
ORDER BY id DESC;
```

### 查看管理员用户

```sql
-- 查看后台管理员
SELECT 
    id,
    username,
    email,
    mobile,
    status,
    department_id,
    customer_id,
    created_at
FROM admin_users
WHERE deleted_at = 0
ORDER BY id DESC;
```

---

## 常见问题排查

### 1. Token 验证失败

**问题**: 返回 "Token无效"

**排查步骤**:
1. 检查 Token 格式是否正确 (格式: `{id}|{plainToken}`)
2. 查询数据库确认 Token 是否存在
3. 检查 Token 的 SHA256 哈希是否匹配

```sql
-- 手动验证 Token
SELECT * FROM personal_access_tokens 
WHERE token = SHA2('your_plain_token_part', 256);
```

### 2. 无权限访问

**问题**: 政工端返回 "无权限访问"

**排查步骤**:
1. 确认 person_type 是否为 2
2. 检查 tokenable_type 是否为 `App\Models\Person`

```sql
-- 检查人员类型
SELECT id, name, person_type FROM persons WHERE id = ?;
```

### 3. 用户已被禁用

**问题**: 返回 "用户已被禁用"

**排查步骤**:
1. 检查 persons.status 或 admin_users.status
2. 确认 status = 1 (正常状态)

```sql
-- 检查用户状态
SELECT id, name, status FROM persons WHERE id = ?;
SELECT id, username, status FROM admin_users WHERE id = ?;
```

### 4. Token 已过期

**问题**: 返回 "Token已过期"

**排查步骤**:
1. 检查 expires_at 字段
2. 重新登录获取新 Token

```sql
-- 检查 Token 过期时间
SELECT 
    id,
    expires_at,
    NOW() as current_time,
    CASE 
        WHEN expires_at IS NULL THEN '永不过期'
        WHEN expires_at > NOW() THEN '有效'
        ELSE '已过期'
    END as status
FROM personal_access_tokens
WHERE token = SHA2('your_plain_token_part', 256);
```

---

## 性能测试

### 使用 Apache Bench

```bash
# 测试政工端接口
ab -n 1000 -c 10 \
  -H "Authorization: Bearer your_token_here" \
  http://localhost:8080/api/staff/info

# 测试学校后台接口
ab -n 1000 -c 10 \
  -H "Authorization: Bearer your_token_here" \
  http://localhost:8080/api/base/info
```

### 使用 wrk

```bash
# 测试政工端接口
wrk -t4 -c100 -d30s \
  -H "Authorization: Bearer your_token_here" \
  http://localhost:8080/api/staff/info
```

---

## 集成测试脚本

### Bash 脚本示例

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"
STAFF_TOKEN="your_staff_token"
ADMIN_TOKEN="your_admin_token"

echo "=== Testing Health Check ==="
curl -s $BASE_URL/health | jq

echo -e "\n=== Testing Staff Info ==="
curl -s -H "Authorization: Bearer $STAFF_TOKEN" \
  $BASE_URL/api/staff/info | jq

echo -e "\n=== Testing Admin Info ==="
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  $BASE_URL/api/base/info | jq

echo -e "\n=== Testing Invalid Token ==="
curl -s -H "Authorization: Bearer invalid_token" \
  $BASE_URL/api/staff/info | jq

echo -e "\n=== Testing No Token ==="
curl -s $BASE_URL/api/staff/info | jq
```

保存为 `test.sh` 并执行:
```bash
chmod +x test.sh
./test.sh
```

---

## 监控和日志

### 查看服务日志

服务启动后会输出日志:
```
Database connected successfully
Server starting on port 8080
```

### Token 使用记录

每次成功验证后，会更新 `last_used_at` 字段:

```sql
-- 查看最近使用的 Token
SELECT 
    tokenable_type,
    tokenable_id,
    last_used_at,
    TIMESTAMPDIFF(MINUTE, last_used_at, NOW()) as minutes_ago
FROM personal_access_tokens
WHERE last_used_at IS NOT NULL
ORDER BY last_used_at DESC
LIMIT 10;
```

---

## 安全测试

### 1. SQL 注入测试

```bash
# 尝试 SQL 注入
curl -H "Authorization: Bearer 1' OR '1'='1" \
  http://localhost:8080/api/staff/info
```

**预期**: 返回 "Token格式错误" (使用参数化查询，安全)

### 2. XSS 测试

```bash
# 尝试 XSS
curl -H "Authorization: Bearer <script>alert('xss')</script>" \
  http://localhost:8080/api/staff/info
```

**预期**: 返回 "Token格式错误" (Token 格式验证，安全)

### 3. 暴力破解测试

```bash
# 尝试多次错误 Token
for i in {1..100}; do
  curl -s -H "Authorization: Bearer fake_token_$i" \
    http://localhost:8080/api/staff/info
done
```

**建议**: 生产环境应添加限流中间件

---

## 总结

本测试文档涵盖了:
- ✅ Token 验证功能测试
- ✅ 错误场景测试
- ✅ 数据库验证
- ✅ 性能测试
- ✅ 安全测试
- ✅ 集成测试脚本
- ✅ 使用场景示例

## ⚠️ 重要提醒

**本服务是 Token 验证中台，不提供：**
- ❌ 登录接口（由 staff_member_api 和 catch_admin_base 提供）
- ❌ 用户注册接口
- ❌ 密码修改接口
- ❌ 用户信息查询接口（只返回 ID）

**服务职责：**
- ✅ 验证 Token 有效性
- ✅ 返回用户身份 ID
- ✅ 检查权限类型（政工端）

确保所有测试通过后再部署到生产环境。
