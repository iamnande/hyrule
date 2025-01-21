package signup

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/rest"
	"github.com/iamnande/hyrule/internal/services/signup"
)

type SignUpService interface {
	Signup(
		ctx context.Context,
		request *signup.SignUpRequest,
	) (*models.User, error)
}

type API struct {
	logger        *slog.Logger
	signUpService SignUpService
}

type Params struct {
	fx.In

	Logger        *slog.Logger
	SignUpService SignUpService
}

type Result struct {
	fx.Out

	APIHandler rest.APIHandler `group:"signup-api:v1:apis"`
}

func NewSignUpAPI(params Params) (Result, error) {
	return Result{
		APIHandler: &API{
			signUpService: params.SignUpService,
		},
	}, nil
}

func (api *API) Handler() http.Handler {
	servicesAPI := chi.NewRouter()
	servicesAPI.Post("/", api.SignUp)
	return servicesAPI
}

func (api *API) URLPath() string {
	return "/signup"
}
