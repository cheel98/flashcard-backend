package jwt

import (
	"time"

	"go.uber.org/fx"
)

// Module JWT模块
var Module = fx.Options(
	fx.Provide(
		NewJWTManager,
	),
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey            string        `json:"secret_key"`
	AccessTokenDuration  time.Duration `json:"access_token_duration"`
	RefreshTokenDuration time.Duration `json:"refresh_token_duration"`
}

// NewJWTManagerFromConfig 从配置创建JWT管理器
func NewJWTManagerFromConfig(config JWTConfig) *JWTManager {
	return NewJWTManager(
		config.SecretKey,
		config.AccessTokenDuration,
		config.RefreshTokenDuration,
	)
}