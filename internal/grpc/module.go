package grpc

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewUserGRPCServer),
	fx.Provide(NewDictionaryGRPCServer),
	fx.Provide(NewFavoriteGRPCServer),
	fx.Provide(NewTranslationServerWithConfig),
	fx.Provide(NewHealthGRPCServer),
)
