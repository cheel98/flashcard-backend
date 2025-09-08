package logger

import (
	"go.uber.org/fx"
)

// Module 日志模块
var Module = fx.Options(
	fx.Provide(NewLogger),
)
