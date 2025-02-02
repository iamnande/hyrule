package config

import (
	"go.uber.org/fx"
)

var RegistrationAPIModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadDatabase()),
	fx.Provide(LoadEmail()),
	fx.Provide(LoadPassword()),
	// TODO: shared invite config? nah. config/service alignment?
	fx.Provide(LoadInvite()),
)

var InviteAPIModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadDatabase()),
)
