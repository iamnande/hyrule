package registration

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/domains/registration"
	"github.com/iamnande/hyrule/internal/rest"
)

type RegistrationDomain interface {
	RegisterNewUser(
		ctx context.Context,
		input *registration.RegisterNewUserInput,
	) (*registration.RegisterNewUserOutput, error)
}

type API struct {
	logger             *slog.Logger
	registrationDomain RegistrationDomain
}

type Params struct {
	fx.In

	Logger             *slog.Logger
	RegistrationDomain RegistrationDomain
}

type Result struct {
	fx.Out

	APIHandler rest.APIHandler `group:"registration-api:v1:apis"`
}

func NewRegistrationAPI(params Params) (Result, error) {
	return Result{
		APIHandler: &API{
			logger:             params.Logger,
			registrationDomain: params.RegistrationDomain,
		},
	}, nil
}

func (api *API) Handler() http.Handler {
	registrationAPI := chi.NewRouter()
	registrationAPI.Post("/", api.RegisterNewUser)
	return registrationAPI
}

func (api *API) URLPath() string {
	return "/register"
}
