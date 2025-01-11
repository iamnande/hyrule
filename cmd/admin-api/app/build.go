package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

func Build() []fx.Option {
	return []fx.Option{
		config.AdminAPIConfigModule,

		// TODO: tracing

		// TODO: service layer

		// TODO: data layer

		// TODO: domain layer

		// TODO: API(s) + routing layer
	}
}
