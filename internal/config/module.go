package config

import (
	"go.uber.org/fx"
)

// Module 配置模块
var Module = fx.Options(
	fx.Provide(LoadConfig),
)
