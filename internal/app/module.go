package app

import (
	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/cheel98/flashcard-backend/internal/database"
	"github.com/cheel98/flashcard-backend/internal/handler"
	"github.com/cheel98/flashcard-backend/internal/middleware"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"github.com/cheel98/flashcard-backend/pkg/logger"
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
	// JWT模块
	jwt.Module,
	// 仓储模块
	repository.Module,
	// 服务模块
	service.Module,
	// 中间件模块
	middleware.Module,
	// 处理器模块
	handler.Module,
	// 服务器模块
	fx.Provide(NewServer),
)
