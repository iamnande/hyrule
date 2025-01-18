package signup

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/repositories/user"
)

type Service struct{}

type SignUpRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func NewService() *Service {
	return &Service{}
}

func (service *Service) Signup(_ context.Context, request *SignUpRequest) (*models.User, error) {
	return models.MarshalUser(&user.Record{
		ID:        uuid.Must(uuid.NewV7()),
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}), nil
}
