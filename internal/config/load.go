package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"

	"github.com/iamnande/hyrule/internal/version"
)

func load[T any](prefix string) func() (T, error) {
	return func() (T, error) {
		cfg := new(T)
		opts := env.Options{
			Prefix: fmt.Sprintf("%s_", strings.ToUpper(version.ServicePrefix)),
		}
		if prefix != "" {
			opts.Prefix = fmt.Sprintf("%s_%s_", strings.ToUpper(version.ServicePrefix), prefix)
		}

		if err := env.ParseWithOptions(cfg, opts); err != nil {
			return *cfg, err
		}

		return *cfg, nil
	}
}
