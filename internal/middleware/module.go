package middleware

import (
	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module 中间件模块
var Module = fx.Options(
	fx.Provide(
		NewAuthMiddleware,
	),
)
