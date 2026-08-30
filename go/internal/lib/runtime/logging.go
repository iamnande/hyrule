package runtime

import (
	"log/slog"
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/iamnande/hyrule/go/internal/lib/config"
	"github.com/iamnande/hyrule/go/internal/lib/version"
)

var LoggingModule = fx.Options(
	fx.Module("logging",
		fx.Provide(NewLogger),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
	),
)

type LoggerParams struct {
	fx.In

	Deployment config.Deployment
	Service    version.ServiceInfo
}

type LoggerResult struct {
	fx.Out

	Logger *slog.Logger
}

func NewLogger(params LoggerParams) (LoggerResult, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger = logger.With(
		slog.Group("deployment",
			slog.String("region", string(params.Deployment.Region)),
			slog.String("environment", string(params.Deployment.Environment)),
		),
		slog.Group("service",
			slog.String("name", params.Service.Name),
			slog.String("version", params.Service.Version),
			slog.String("commit", params.Service.Commit),
		),
	)
	slog.SetDefault(logger)
	return LoggerResult{
		Logger: logger,
	}, nil
}
