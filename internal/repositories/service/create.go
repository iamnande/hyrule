package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateServiceRequest struct {
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (repository *Repository) CreateService(
	ctx context.Context,
	request *CreateServiceRequest,
) (*Record, error) {
	return &Record{
		ID:          uuid.New(),
		Name:        request.Name,
		Description: request.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
