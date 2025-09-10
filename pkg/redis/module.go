package redis

import (
	"go.uber.org/fx"
)

// Module Redis模块
var Module = fx.Options(
	fx.Provide(
		NewRedisClient,
	),
)
