package app

import (
	"log/slog"
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/version"
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
}

type LoggerResult struct {
	fx.Out

	Logger *slog.Logger
}

func NewLogger(params LoggerParams) (LoggerResult, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger = logger.With(
		slog.Group("runtime",
			slog.String("region", string(params.Deployment.Region)),
			slog.String("environment", string(params.Deployment.Environment)),
		),
		slog.Group("service",
			slog.String("name", version.ServiceName),
			slog.String("version", version.ServiceVersion),
			slog.String("commit", version.ServiceCommit),
		),
	)
	slog.SetDefault(logger)
	return LoggerResult{
		Logger: logger,
	}, nil
}
