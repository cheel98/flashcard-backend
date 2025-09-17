package app

import (
	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/cheel98/flashcard-backend/internal/database"
	"github.com/cheel98/flashcard-backend/internal/grpc"
	"github.com/cheel98/flashcard-backend/internal/handler"
	"github.com/cheel98/flashcard-backend/internal/middleware"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/pkg"
	"go.uber.org/fx"
)

// Module 定义应用的依赖注入模块
var Module = fx.Options(
	// 配置模块
	config.Module,
	// 数据库模块
	database.Module,
	// 仓储模块
	repository.Module,
	// 中间件模块
	middleware.Module,
	// 处理器模块
	handler.Module,
	pkg.Module,
	// 服务器模块
	grpc.Module,
	fx.Provide(NewServer),
)
