package database

import (
	"fmt"

	"flashcard-backend/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database 数据库结构体
type Database struct {
	DB     *gorm.DB
	logger *zap.Logger
}

// NewDatabase 创建新的数据库连接
func NewDatabase(cfg *config.Config, logger *zap.Logger) (*Database, error) {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone,
	)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	// 自动迁移数据库表
	if err := db.AutoMigrate(
	//&model.Deck{},
	// 配置orm对象自动迁移数据库
	); err != nil {
		logger.Error("Failed to migrate database", zap.Error(err))
		return nil, err
	}

	logger.Info("Database connected successfully")

	return &Database{
		DB:     db,
		logger: logger,
	}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
