package app

import (
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/invite-api/health"
	v1InviteRouter "github.com/iamnande/hyrule/internal/apis/invite-api/v1"
	v1InviteAPI "github.com/iamnande/hyrule/internal/apis/invite-api/v1/invites"
	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/database"
)

func Build() []fx.Option {
	return []fx.Option{
		config.InviteAPIModule,

		// service layer
		fx.Provide(
		// TODO: invite service (AcceptInvite)
		// fx.Annotate(invite.NewService, fx.As(new(inviteRepository.InviteService))),
		),

		// data layer
		fx.Provide(
			// database client
			fx.Annotate(database.NewDatabaseClient,
				fx.As(new(healthAPI.DatabaseClient)),
			),
		),

		// domain layer
		fx.Provide(
		// fx.Annotate(inviteDomain.NewDomain, fx.As(new(v1InviteAPI.InviteDomain))),
		),

		// runtime
		fx.Provide(
			healthAPI.NewHealthAPI,
			v1InviteAPI.NewInviteAPI,
			v1InviteRouter.NewInviteAPIRouter,
		),
	}
}
