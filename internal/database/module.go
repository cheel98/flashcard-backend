package database

import (
	"go.uber.org/fx"
)

// Module 数据库模块
var Module = fx.Options(
	fx.Provide(NewDatabase),
)
