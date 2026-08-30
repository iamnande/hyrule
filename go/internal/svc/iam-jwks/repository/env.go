package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/iamnande/hyrule/go/internal/lib/version"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/domain"
)

type EnvConfig struct {
	Keys string `env:"KEYS"`
}

func LoadEnvConfig() (EnvConfig, error) {
	cfg := EnvConfig{}
	opts := env.Options{
		Prefix: fmt.Sprintf("%s_IAM_JWKS_", strings.ToUpper(version.ServicePrefix)),
	}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return cfg, err
	}
	return cfg, nil
}

type EnvRepository struct {
	keys string
}

func NewEnv(cfg EnvConfig) *EnvRepository {
	return &EnvRepository{keys: cfg.Keys}
}

type envKey struct {
	ID        string    `json:"id"`
	Algorithm string    `json:"algorithm"`
	PublicKey string    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *EnvRepository) List(_ context.Context) ([]domain.Key, error) {
	var envKeys []envKey
	if err := json.Unmarshal([]byte(r.keys), &envKeys); err != nil {
		return nil, fmt.Errorf("parse keys: %w", err)
	}
	keys := make([]domain.Key, len(envKeys))
	for i, ek := range envKeys {
		keys[i] = domain.Key{
			ID:        ek.ID,
			Algorithm: ek.Algorithm,
			PublicKey: ek.PublicKey,
			CreatedAt: ek.CreatedAt,
		}
	}
	return keys, nil
}
