package domain

import (
	"context"
	"time"
)

type Key struct {
	ID        string
	Algorithm string
	PublicKey string
	CreatedAt time.Time
}

type keyStore interface {
	List(ctx context.Context) ([]Key, error)
}

type Service struct {
	store keyStore
}

func NewService(store keyStore) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) ([]Key, error) {
	return s.store.List(ctx)
}
