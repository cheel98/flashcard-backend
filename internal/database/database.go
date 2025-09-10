package database

import (
	"fmt"
	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/cheel98/flashcard-backend/internal/model"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database 数据库结构体
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建新的数据库连接
func NewDatabase(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
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
		&model.User{},
		&model.UserSettings{},
		&model.UserPreferences{},
		&model.UserLogs{},
		&model.PaymentRecord{},
		&model.Favorite{},
		&model.StudyRecord{},
		&model.Dictionary{},
		&model.DictionaryAudio{},
		&model.DictionaryMetadata{},
	); err != nil {
		logger.Error("Failed to migrate database", zap.Error(err))
		return nil, err
	}

	logger.Info("Database connected successfully")

	return db, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
