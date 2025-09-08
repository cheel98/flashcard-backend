package service

import "go.uber.org/fx"

// Module 定义服务层的依赖注入模块
var Module = fx.Options(
	// 用户服务
	fx.Provide(NewUserService),
	// 词典服务
	fx.Provide(NewDictionaryService),
	// 收藏服务
	fx.Provide(NewFavoriteService),
)
