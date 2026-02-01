package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config 数据库配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Charset  string
}

// NewMySQLConnection 创建MySQL连接
func NewMySQLConnection(cfg Config) (*sql.DB, error) {
	if cfg.Charset == "" {
		cfg.Charset = "utf8mb4"
	}

	// 添加连接参数优化
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 优化连接池参数
	db.SetMaxOpenConns(50)                 // 最大打开连接数（降低以避免耗尽）
	db.SetMaxIdleConns(25)                 // 最大空闲连接数（增加以保持连接）
	db.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期（缩短以避免超时）
	db.SetConnMaxIdleTime(3 * time.Minute) // 连接最大空闲时间（缩短以保持活跃）

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
