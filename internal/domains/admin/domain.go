package admin

import (
	"context"

	serviceRepository "github.com/iamnande/hyrule/internal/repositories/service"
	"go.uber.org/fx"
)

type ServiceRepository interface {
	CreateService(
		ctx context.Context,
		request *serviceRepository.CreateServiceRequest,
	) (*serviceRepository.Record, error)
}

type Domain struct {
	serviceRepository ServiceRepository
}

type Params struct {
	fx.In

	ServiceRepository ServiceRepository
}

func NewDomain(params Params) *Domain {
	return &Domain{
		serviceRepository: params.ServiceRepository,
	}
}
