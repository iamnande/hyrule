package runtime

import (
	"errors"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/config"
	"github.com/iamnande/hyrule/go/internal/lib/tracing"
	"github.com/iamnande/hyrule/go/internal/lib/version"
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
	Service    version.ServiceInfo
}

func StartTracer(params TracerParams) error {
	if err := tracing.Init(tracing.InitOptions{
		Enabled:      params.Config.Enabled,
		SampleRate:   params.Config.SampleRate,
		IngestionURL: params.Config.IngestionURL,
		Release:      params.Service.Version,
		Environment:  string(params.Deployment.Environment),
		Tags: map[string]string{
			"service.commit":         params.Service.Commit,
			"service.name":           params.Service.Name,
			"service.version":        params.Service.Version,
			"deployment.environment": string(params.Deployment.Environment),
			"deployment.region":      string(params.Deployment.Region),
		},
	}); err != nil {
		return errors.Join(ErrFailedToStartTracer, err)
	}

	return nil
}
