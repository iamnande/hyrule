package invites

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/rest"
)

type API struct {
	logger *slog.Logger
}

type Params struct {
	fx.In

	Logger *slog.Logger
}

type Result struct {
	fx.Out

	APIHandler rest.APIHandler `group:"invite-api:v1:apis"`
}

func NewInviteAPI(params Params) (Result, error) {
	return Result{
		APIHandler: &API{
			logger: params.Logger,
		},
	}, nil
}

func (api *API) Handler() http.Handler {
	inviteAPI := chi.NewRouter()
	inviteAPI.Get("/{token}/accept", api.AcceptInvite)
	return inviteAPI
}

func (api *API) URLPath() string {
	return "/invites"
}
