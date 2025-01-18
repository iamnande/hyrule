package app

import (
	"errors"

	"go.uber.org/fx"

	"github.com/getsentry/sentry-go"

	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/version"
)

var TracingModule = fx.Module("tracing",
	fx.Invoke(StartTracer),
)

var (
	ErrFailedToStartTracer = errors.New("failed to start tracer")
)

type TracerParams struct {
	fx.In

	Config     config.Tracing
	Deployment config.Deployment
}

func StartTracer(params TracerParams) error {
	if !params.Config.Enabled {
		return nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    params.Config.Enabled,
		TracesSampleRate: params.Config.SampleRate,
		Dsn:              params.Config.IngestionURL,
		Release:          version.ServiceVersion,
		Environment:      string(params.Deployment.Environment),
		Tags: map[string]string{
			"service.commit":         version.ServiceCommit,
			"service.name":           version.ServiceName,
			"service.version":        version.ServiceVersion,
			"deployment.environment": string(params.Deployment.Environment),
			"deployment.region":      string(params.Deployment.Region),
		},
	}); err != nil {
		return errors.Join(ErrFailedToStartTracer, err)
	}

	return nil
}
