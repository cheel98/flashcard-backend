package logger

import (
	"flashcard-backend/internal/config"
	"go.uber.org/zap"
)

// NewLogger 创建新的日志实例
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	// 根据环境配置选择日志配置
	if cfg.Server.Env == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	switch cfg.Logger.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// 设置日志格式
	if cfg.Logger.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// 构建日志实例
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	// 设置全局日志实例
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// Sync 同步日志缓冲区
func Sync(logger *zap.Logger) {
	_ = logger.Sync()
}
