package app

import (
	"flashcard-backend/internal/config"
	"flashcard-backend/internal/database"
	"flashcard-backend/internal/handler"
	"flashcard-backend/internal/service"
	"flashcard-backend/pkg/logger"
	"go.uber.org/fx"
)

// Module 定义应用的依赖注入模块
var Module = fx.Options(
	// 配置模块
	config.Module,
	// 日志模块
	logger.Module,
	// 数据库模块
	database.Module,
	// 服务模块
	service.Module,
	// 处理器模块
	handler.Module,
	// 服务器模块
	fx.Provide(NewServer),
)
