package module

import (
	"go.uber.org/fx"
	"med_book/internal/config"
)

var ConfigModule = fx.Module("config",
	fx.Provide(config.Load)
)