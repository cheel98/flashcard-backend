package repository

import (
	"go.uber.org/fx"
)

// Module repository模块
var Module = fx.Options(
	fx.Provide(NewDictionaryRepository),
	fx.Provide(NewUserRepository),
	fx.Provide(NewFavoriteRepository),
)
