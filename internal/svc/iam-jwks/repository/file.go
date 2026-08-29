package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/iamnande/hyrule/internal/lib/version"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/domain"
)

type FileConfig struct {
	Path string `env:"KEYS_FILE_PATH"`
}

func LoadFileConfig() (FileConfig, error) {
	cfg := FileConfig{}
	opts := env.Options{
		Prefix: fmt.Sprintf("%s_IAM_JWKS_", strings.ToUpper(version.ServicePrefix)),
	}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return cfg, err
	}
	return cfg, nil
}

type FileRepository struct {
	path string
}

func NewFile(cfg FileConfig) *FileRepository {
	return &FileRepository{path: cfg.Path}
}

type fileKey struct {
	ID        string    `json:"id"`
	Algorithm string    `json:"algorithm"`
	PublicKey string    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *FileRepository) List(_ context.Context) ([]domain.Key, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("read keys file %q: %w", r.path, err)
	}
	var fileKeys []fileKey
	if err := json.Unmarshal(data, &fileKeys); err != nil {
		return nil, fmt.Errorf("parse keys file %q: %w", r.path, err)
	}
	keys := make([]domain.Key, len(fileKeys))
	for i, fk := range fileKeys {
		keys[i] = domain.Key{
			ID:        fk.ID,
			Algorithm: fk.Algorithm,
			PublicKey: fk.PublicKey,
			CreatedAt: fk.CreatedAt,
		}
	}
	return keys, nil
}
