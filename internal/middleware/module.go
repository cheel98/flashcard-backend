package middleware

import (
	"go.uber.org/fx"
)

// Module 中间件模块
var Module = fx.Options(
	fx.Provide(
		NewAuthMiddleware,
	),
)
