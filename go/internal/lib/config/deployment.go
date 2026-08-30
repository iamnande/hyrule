package config

import (
	"fmt"
	"slices"
)

type Deployment struct {
	Region      Region      `env:"REGION" envDefault:"us-east-2"`
	Environment Environment `env:"ENVIRONMENT" envDefault:"local"`
}

func LoadDeployment() func() (Deployment, error) {
	loader := Load[Deployment]("")
	return func() (Deployment, error) {
		cfg, err := loader()
		if err != nil {
			return cfg, err
		}
		if !slices.Contains(validRegions, cfg.Region) {
			return cfg, fmt.Errorf("invalid region %q", cfg.Region)
		}
		if !slices.Contains(validEnvironments, cfg.Environment) {
			return cfg, fmt.Errorf("invalid environment %q", cfg.Environment)
		}
		return cfg, nil
	}
}
