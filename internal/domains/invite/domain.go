package invite

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/services/invite"
)

type InviteService interface {
	AcceptInvite(
		ctx context.Context,
		input *invite.AcceptInviteRequest,
	) error
}

type Domain struct {
	logger *slog.Logger

	inviteService InviteService
}

type Params struct {
	fx.In

	Logger        *slog.Logger
	InviteService InviteService
}

func NewDomain(params Params) *Domain {
	return &Domain{
		logger:        params.Logger,
		inviteService: params.InviteService,
	}
}
