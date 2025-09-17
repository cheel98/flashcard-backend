package jwt

import (
	"github.com/cheel98/flashcard-backend/internal/config"
	"time"

	"go.uber.org/fx"
)

// Module JWT模块
var Module = fx.Options(
	fx.Provide(
		NewJWTManagerFromConfig,
	),
)

// NewJWTManagerFromConfig 从配置创建JWT管理器
func NewJWTManagerFromConfig(config *config.Config) *JWTManager {
	return NewJWTManager(
		config.JWT.SecretKey,
		time.Duration(config.JWT.AccessTokenDuration)*time.Minute,
		time.Duration(config.JWT.RefreshTokenDuration)*time.Hour,
	)
}
