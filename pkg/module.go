package pkg

import (
	"github.com/cheel98/flashcard-backend/pkg/email"
	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"github.com/cheel98/flashcard-backend/pkg/logger"
	"github.com/cheel98/flashcard-backend/pkg/redis"
	"go.uber.org/fx"
)

// Module Redis模块
var Module = fx.Options(
	email.Module,
	jwt.Module,
	logger.Module,
	redis.Module,
)
