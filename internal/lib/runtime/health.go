package runtime

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/lib/rest"
	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	"github.com/iamnande/hyrule/internal/lib/version"
)

type HealthAPIResult struct {
	fx.Out

	API rest.APIHandler `name:"health"`
}

type HealthAPIParams struct {
	fx.In

	Logger       *slog.Logger
	Deployment   config.Deployment
	Service      version.ServiceInfo
	Probes       health.Probes
	Dependencies []health.Option `group:"healthDependencies"`
}

// NewHealthAPI wires the health capability every service gets for free -
// only the Probes (and any Dependencies) are the service's own (see
// internal/lib/rest/capabilities/health).
func NewHealthAPI(params HealthAPIParams) (HealthAPIResult, error) {
	opts := append([]health.Option{
		health.WithLogger(params.Logger),
		health.WithServiceMetadata(&health.ServiceMetadata{
			Name:        params.Service.Name,
			Version:     params.Service.Version,
			Commit:      params.Service.Commit,
			Environment: params.Deployment.Environment,
			Region:      params.Deployment.Region,
			Links: map[string]string{
				"repository":  version.RepositoryURL,
				"docs":        version.RepositoryURL + "/blob/main/docs/architecture.md",
				"conventions": version.RepositoryURL + "/blob/main/docs/conventions.md",
			},
		}),
	}, params.Dependencies...)

	register, err := health.NewAPI(params.Probes, opts...)
	if err != nil {
		return HealthAPIResult{}, err
	}
	return HealthAPIResult{
		API: rest.APIHandler(register),
	}, nil
}
