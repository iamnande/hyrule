package admin

import (
	"context"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/repositories/service"
)

type CreateServiceRequest struct {
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r CreateServiceRequest) Validate() error {
	// TODO: native?
	return nil
}

func (domain *Domain) CreateService(
	ctx context.Context,
	request *CreateServiceRequest,
) (*models.Service, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	record, err := domain.serviceRepository.CreateService(ctx, &service.CreateServiceRequest{
		Domain:      request.Domain,
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		return nil, err
	}

	return models.MarshalService(&service.Record{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}), nil
}
