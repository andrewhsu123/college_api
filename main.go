package main

import (
	"log"
	"os"
	"fmt"
	"college_api/internal/handler/base"
	"college_api/internal/handler/staff"
	baseMiddleware "college_api/internal/middleware/base"
	staffMiddleware "college_api/internal/middleware/staff"
	"college_api/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 数据库配置
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnvAsInt("DB_PORT", 3306),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "root"),
		Database: getEnv("DB_DATABASE", "college_db_base"),
		Charset:  "utf8mb4",
	}

	// 连接数据库
	db, err := database.NewMySQLConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// 创建Gin引擎
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "College API Service",
		})
	})

	// 政工端路由组
	staffGroup := r.Group("/api/staff")
	{
		// 创建中间件和处理器
		staffAuth := staffMiddleware.NewAuthMiddleware(db)
		staffHandler := staff.NewAuthHandler()

		// 需要认证的路由
		staffGroup.Use(staffAuth.Authenticate())
		{
			staffGroup.GET("/info", staffHandler.GetPersonInfo)
			// 在这里添加更多政工端路由
		}
	}

	// 学校后台路由组
	baseGroup := r.Group("/api/base")
	{
		// 创建中间件和处理器
		baseAuth := baseMiddleware.NewAuthMiddleware(db)
		baseHandler := base.NewAuthHandler()

		// 需要认证的路由
		baseGroup.Use(baseAuth.Authenticate())
		{
			baseGroup.GET("/info", baseHandler.GetUserInfo)
			// 在这里添加更多学校后台路由
		}
	}

	// 启动服务
	port := getEnv("PORT", "8081")
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}
