package health

import (
	"log/slog"
	"time"
)

type config struct {
	logger           *slog.Logger
	timeout          time.Duration
	serviceMetadata  *ServiceMetadata
	dependencyChecks []dependencyCheck
}

type Option func(*config)

func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *config) {
		cfg.timeout = timeout
	}
}

func WithServiceMetadata(metadata *ServiceMetadata) Option {
	return func(cfg *config) {
		cfg.serviceMetadata = metadata
	}
}

func WithHardDependency(name string, check DependencyCheckFn) Option {
	return func(cfg *config) {
		cfg.dependencyChecks = append(cfg.dependencyChecks, dependencyCheck{
			name:      name,
			checkType: DependencyCheckTypeHard,
			check:     check,
		})
	}
}

func WithSoftDependency(name string, check DependencyCheckFn) Option {
	return func(cfg *config) {
		cfg.dependencyChecks = append(cfg.dependencyChecks, dependencyCheck{
			name:      name,
			checkType: DependencyCheckTypeSoft,
			check:     check,
		})
	}
}
