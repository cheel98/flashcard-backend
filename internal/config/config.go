package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构体
type Config struct {
	Server         ServerConfig   `json:"server"`
	Database       DatabaseConfig `json:"database"`
	Logger         LoggerConfig   `json:"logger"`
	JWT            JWTConfig      `json:"jwt"`
	Redis          RedisConfig    `json:"redis"`
	Email          EmailConfig    `json:"email"`
	TransferConfig TransferConfig `json:"transfer_config"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port     int    `json:"port"`      // gRPC 端口
	HTTPPort int    `json:"http_port"` // HTTP 端口 (gRPC-Gateway)
	Host     string `json:"host"`
	Env      string `json:"env"`
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

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
}
type Engine string

const (
	YOUDAO Engine = "Youdao"
)

// 翻译引擎设置
type TransferConfig struct {
	URL       string `json:"url"`
	Engine    Engine `json:"engine"`
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件（如果存在）
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:     getEnvAsInt("SERVER_PORT", 8080),
			HTTPPort: getEnvAsInt("HTTP_PORT", 8081),
			Host:     getEnv("SERVER_HOST", "localhost"),
			Env:      getEnv("APP_ENV", "development"),
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
			AccessTokenDuration:  getEnvAsInt("JWT_ACCESS_TOKEN_DURATION", 15),   // 15分钟
			RefreshTokenDuration: getEnvAsInt("JWT_REFRESH_TOKEN_DURATION", 168), // 7天
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", ""),
			FromName:     getEnv("FROM_NAME", "Flashcard App"),
		},
		TransferConfig: TransferConfig{
			URL:       getEnv("TRANSFER_URL", "https://openapi.youdao.com/api"),
			Engine:    YOUDAO,
			AppKey:    getEnv("APP_KEY", ""),
			AppSecret: getEnv("APP_SECRET", ""),
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
