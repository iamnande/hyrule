package services

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/domains/admin"
	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/rest"
)

type AdminDomain interface {
	CreateService(
		ctx context.Context,
		request *admin.CreateServiceRequest,
	) (*models.Service, error)
}

type API struct {
	logger      *slog.Logger
	adminDomain AdminDomain
}

type Params struct {
	fx.In

	Logger      *slog.Logger
	AdminDomain AdminDomain
}

type Result struct {
	fx.Out

	APIHandler rest.APIHandler `group:"admin-api:v1:apis"`
}

func Build(params Params) (Result, error) {
	return Result{
		APIHandler: &API{
			adminDomain: params.AdminDomain,
		},
	}, nil
}

func (api *API) Handler() http.Handler {
	servicesAPI := chi.NewRouter()
	servicesAPI.Post("/", api.CreateService)
	return servicesAPI
}

func (api *API) URLPath() string {
	return "/services"
}
