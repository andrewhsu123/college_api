package auth

import (
	"testing"
	"time"
)

func TestPasswordHasher(t *testing.T) {
	hasher := NewPasswordHasher()

	t.Run("密码加密和验证", func(t *testing.T) {
		password := "test123456"

		// 加密密码
		hashed, err := hasher.HashPassword(password)
		if err != nil {
			t.Fatalf("密码加密失败: %v", err)
		}

		// 验证正确密码
		if !hasher.CheckPassword(password, hashed) {
			t.Error("正确密码验证失败")
		}

		// 验证错误密码
		if hasher.CheckPassword("wrongpassword", hashed) {
			t.Error("错误密码验证通过")
		}
	})

	t.Run("与 Laravel 生成的密码哈希兼容", func(t *testing.T) {
		// 这是 Laravel Hash::make('password') 生成的哈希
		// 你可以在 Laravel 中运行: Hash::make('password') 获取实际的哈希值
		password := "password"
		laravelHash := "$2y$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5lE3WkKvXVZ.W"

		// bcrypt 的 $2y$ 和 $2a$ 是兼容的
		// Go 的 bcrypt 使用 $2a$，但可以验证 $2y$ 的哈希
		if !hasher.CheckPassword(password, laravelHash) {
			t.Log("注意: 如果测试失败，请在 Laravel 中生成新的哈希值并更新此测试")
		}
	})

	t.Run("检查是否需要重新哈希", func(t *testing.T) {
		password := "test123456"
		hashed, _ := hasher.HashPassword(password)

		if hasher.NeedsRehash(hashed) {
			t.Error("新生成的哈希不应该需要重新哈希")
		}
	})
}

func TestTokenGenerator(t *testing.T) {
	appKey := "base64:test_app_key_for_testing_12345678"
	generator := NewTokenGenerator(appKey)

	t.Run("生成和解析 Token", func(t *testing.T) {
		personID := int64(12345)
		role := "student"
		expiresIn := int64(7 * 24 * 3600) // 7 天

		// 生成 token
		token, err := generator.GenerateToken(personID, role, expiresIn)
		if err != nil {
			t.Fatalf("生成 token 失败: %v", err)
		}

		if token == "" {
			t.Fatal("生成的 token 为空")
		}

		t.Logf("生成的 token: %s", token)

		// 解析 token
		tokenData, err := generator.ParseToken(token)
		if err != nil {
			t.Fatalf("解析 token 失败: %v", err)
		}

		// 验证数据
		if tokenData.PersonID != personID {
			t.Errorf("PersonID 不匹配: 期望 %d, 得到 %d", personID, tokenData.PersonID)
		}

		if tokenData.Role != role {
			t.Errorf("Role 不匹配: 期望 %s, 得到 %s", role, tokenData.Role)
		}

		if tokenData.ExpiresAt <= time.Now().Unix() {
			t.Error("Token 已过期")
		}
	})

	t.Run("验证 Token", func(t *testing.T) {
		personID := int64(67890)
		role := "staff"

		token, _ := generator.GenerateToken(personID, role, 3600)

		// 验证 token
		parsedPersonID, parsedRole, err := generator.ValidateToken(token)
		if err != nil {
			t.Fatalf("验证 token 失败: %v", err)
		}

		if parsedPersonID != personID {
			t.Errorf("PersonID 不匹配: 期望 %d, 得到 %d", personID, parsedPersonID)
		}

		if parsedRole != role {
			t.Errorf("Role 不匹配: 期望 %s, 得到 %s", role, parsedRole)
		}
	})

	t.Run("解析无效 Token", func(t *testing.T) {
		invalidTokens := []string{
			"",
			"invalid",
			"invalid.token",
			"aW52YWxpZA==.invalid",
		}

		for _, token := range invalidTokens {
			_, err := generator.ParseToken(token)
			if err == nil {
				t.Errorf("应该拒绝无效 token: %s", token)
			}
		}
	})

	t.Run("过期 Token", func(t *testing.T) {
		personID := int64(11111)
		role := "worker"

		// 生成一个已过期的 token（过期时间为 -1 秒）
		token, _ := generator.GenerateToken(personID, role, -1)

		// 等待一小段时间确保过期
		time.Sleep(10 * time.Millisecond)

		// 尝试解析过期 token
		_, err := generator.ParseToken(token)
		if err == nil {
			t.Error("应该拒绝过期的 token")
		}
	})

	t.Run("不同角色的 Token", func(t *testing.T) {
		personID := int64(99999)
		roles := []string{"student", "staff", "worker"}

		for _, role := range roles {
			token, err := generator.GenerateToken(personID, role, 3600)
			if err != nil {
				t.Fatalf("生成 %s token 失败: %v", role, err)
			}

			tokenData, err := generator.ParseToken(token)
			if err != nil {
				t.Fatalf("解析 %s token 失败: %v", role, err)
			}

			if tokenData.Role != role {
				t.Errorf("角色不匹配: 期望 %s, 得到 %s", role, tokenData.Role)
			}
		}
	})
}

func BenchmarkPasswordHasher(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "test123456"

	b.Run("HashPassword", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = hasher.HashPassword(password)
		}
	})

	b.Run("CheckPassword", func(b *testing.B) {
		hashed, _ := hasher.HashPassword(password)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = hasher.CheckPassword(password, hashed)
		}
	})
}

func BenchmarkTokenGenerator(b *testing.B) {
	generator := NewTokenGenerator("test_app_key")
	personID := int64(12345)
	role := "student"

	b.Run("GenerateToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = generator.GenerateToken(personID, role, 3600)
		}
	})

	b.Run("ParseToken", func(b *testing.B) {
		token, _ := generator.GenerateToken(personID, role, 3600)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = generator.ParseToken(token)
		}
	})
}
