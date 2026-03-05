package main

import (
	"college_api/internal/handler/base"
	"college_api/internal/handler/open"
	"college_api/internal/handler/staff"
	baseMiddleware "college_api/internal/middleware/base"
	openMiddleware "college_api/internal/middleware/open"
	staffMiddleware "college_api/internal/middleware/staff"
	"college_api/internal/repository"
	"college_api/internal/service"
	"college_api/pkg/database"
	"fmt"
	"log"
	"os"

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
		Host:     getEnv("DB_HOST", "localhost"),
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

	// log.Println("Database connected successfully")

	// 创建仓库和服务
	deptRepo := repository.NewDepartmentRepository(db)
	deptService := service.NewDepartmentService(deptRepo)

	personRepo := repository.NewPersonRepository(db)
	personService := service.NewPersonService(personRepo, deptRepo)

	appRepo := repository.NewApplicationRepository(db)
	appService := service.NewApplicationService(appRepo)

	openPersonRepo := repository.NewOpenPersonRepository(db)
	openPersonService := service.NewOpenPersonService(openPersonRepo)

	collegeRepo := repository.NewCollegeRepository(db)
	collegeService := service.NewCollegeService(collegeRepo)

	// 创建Gin引擎
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "College API Service",
		})
	})

	// 政工端路由组
	staffGroup := r.Group("/api/staff")
	{
		// 创建中间件和处理器
		staffAuth := staffMiddleware.NewAuthMiddleware(db)
		staffAuthHandler := staff.NewAuthHandler(personService)
		staffDeptHandler := staff.NewDepartmentHandler(deptService)
		staffPersonHandler := staff.NewPersonHandler(personService)

		// 需要认证的路由
		staffGroup.Use(staffAuth.Authenticate())
		{
			staffGroup.GET("/info", staffAuthHandler.GetPersonInfo)
			staffGroup.GET("/departments/tree", staffDeptHandler.GetDepartmentTree)
			staffGroup.GET("/departments/list", staffDeptHandler.GetDepartmentList)
			staffGroup.GET("/persons/list", staffPersonHandler.GetPersonList)
		}
	}

	// 学校后台路由组
	baseGroup := r.Group("/api/base")
	{
		// 创建中间件和处理器
		baseAuth := baseMiddleware.NewAuthMiddleware(db)
		baseAuthHandler := base.NewAuthHandler(personService)
		baseDeptHandler := base.NewDepartmentHandler(deptService)
		basePersonHandler := base.NewPersonHandler(personService)

		// 需要认证的路由
		baseGroup.Use(baseAuth.Authenticate())
		{
			baseGroup.GET("/info", baseAuthHandler.GetUserInfo)
			baseGroup.GET("/departments/tree", baseDeptHandler.GetDepartmentTree)
			baseGroup.GET("/departments/list", baseDeptHandler.GetDepartmentList)
			baseGroup.GET("/persons/list", basePersonHandler.GetPersonList)
		}
	}

	// 开放接口路由组
	openGroup := r.Group("/api/open")
	{
		// 创建处理器
		openAppHandler := open.NewApplicationHandler(appService)
		openPersonHandler := open.NewPersonHandler(openPersonService)
		openCollegeHandler := open.NewCollegeHandler(collegeService)

		// 使用秘钥认证中间件
		openGroup.Use(openMiddleware.AuthMiddleware())
		{
			openGroup.GET("/applications/list", openAppHandler.GetApplicationList)
			openGroup.GET("/applications/visible", openAppHandler.GetVisibleApplications)
			openGroup.GET("/staff/list", openPersonHandler.GetStaffList)
			openGroup.GET("/students/list", openPersonHandler.GetStudentList)
			openGroup.GET("/roles/list", openPersonHandler.GetRoleList)
			openGroup.GET("/persons/managers", openPersonHandler.GetManagePersons)
			openGroup.GET("/roles/persons", openPersonHandler.GetRolePersons)
			openGroup.GET("/colleges/list", openCollegeHandler.GetCollegeList)
			openGroup.GET("/campus-areas/list", openCollegeHandler.GetCampusAreaList)
			openGroup.GET("/departments/list", openCollegeHandler.GetDepartmentList)
			openGroup.GET("/staff/by-org", openPersonHandler.GetStaffByOrg)
			openGroup.GET("/students/by-org", openPersonHandler.GetStudentByOrg)
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
