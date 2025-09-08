package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Logger   LoggerConfig   `json:"logger"`
	JWT      JWTConfig      `json:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
	Env  string `json:"env"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
	TimeZone string `json:"time_zone"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey            string `json:"secret_key"`
	AccessTokenDuration  int    `json:"access_token_duration"`  // 分钟
	RefreshTokenDuration int    `json:"refresh_token_duration"` // 小时
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件（如果存在）
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "localhost"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "flashcard_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		JWT: JWTConfig{
			SecretKey:            getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			AccessTokenDuration:  getEnvAsInt("JWT_ACCESS_TOKEN_DURATION", 15),  // 15分钟
			RefreshTokenDuration: getEnvAsInt("JWT_REFRESH_TOKEN_DURATION", 168), // 7天
		},
	}

	return config, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
