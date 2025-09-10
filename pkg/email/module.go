package email

import (
	"go.uber.org/fx"
)

// Module 邮件模块
var Module = fx.Options(
	fx.Provide(
		NewEmailService,
	),
)
