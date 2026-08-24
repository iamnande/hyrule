package domain

import (
	"context"
	"time"
)

type Kind string

const (
	KindHost    Kind = "host"
	KindService Kind = "service"
	KindApp     Kind = "app"
)

type State string

const (
	StateUp    State = "up"
	StateStale State = "stale"
)

type Ping struct {
	Name        string
	Kind        Kind
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	State       State
}

type pingStore interface {
	Upsert(ctx context.Context, name string, kind Kind) (Ping, error)
	List(ctx context.Context) ([]Ping, error)
}

type Service struct {
	store      pingStore
	staleAfter time.Duration
}

func NewService(store pingStore, cfg Config) *Service {
	return &Service{store: store, staleAfter: cfg.StaleAfter}
}

func (s *Service) Record(ctx context.Context, name string, kind Kind) (Ping, error) {
	ping, err := s.store.Upsert(ctx, name, kind)
	if err != nil {
		return Ping{}, err
	}
	return s.withState(ping), nil
}

func (s *Service) List(ctx context.Context) ([]Ping, error) {
	pings, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	for i, ping := range pings {
		pings[i] = s.withState(ping)
	}
	return pings, nil
}

func (s *Service) withState(ping Ping) Ping {
	ping.State = StateUp
	if time.Since(ping.LastSeenAt) > s.staleAfter {
		ping.State = StateStale
	}
	return ping
}
