package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/iamnande/hyrule/internal/lib/version"
)

type Config struct {
	// StaleAfter is how long since a name's last ping before it reads as
	// State: stale instead of up.
	StaleAfter time.Duration `env:"STALE_AFTER" envDefault:"5m"`
}

func LoadConfig() (Config, error) {
	cfg := Config{}
	opts := env.Options{
		Prefix: fmt.Sprintf("%s_PINGS_", strings.ToUpper(version.ServicePrefix)),
	}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return cfg, err
	}
	return cfg, nil
}
