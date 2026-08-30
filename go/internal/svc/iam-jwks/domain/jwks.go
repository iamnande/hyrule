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

type KeySet struct {
	store keyStore
}

func NewKeySet(store keyStore) *KeySet {
	return &KeySet{store: store}
}

func (s *KeySet) List(ctx context.Context) ([]Key, error) {
	return s.store.List(ctx)
}
