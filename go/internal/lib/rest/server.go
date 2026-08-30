package rest

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/config"
)

type server struct {
	*http.Server

	logger *slog.Logger
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger *slog.Logger
	Config config.HTTPServer
	API    http.Handler
}

func NewServer(params Params) *http.Server {
	srvr := &server{
		logger: params.Logger,
		Server: &http.Server{
			Addr:              params.Config.Addr,
			Handler:           params.API,
			ReadHeaderTimeout: params.Config.ReadHeaderTimeout,
			ReadTimeout:       params.Config.ReadTimeout,
			WriteTimeout:      params.Config.WriteTimeout,
			IdleTimeout:       params.Config.IdleTimeout,
		},
	}
	params.Lifecycle.Append(fx.Hook{
		OnStart: srvr.start,
		OnStop:  srvr.stop,
	})
	return srvr.Server
}

func (srvr *server) start(_ context.Context) error {
	lis, err := net.Listen("tcp", srvr.Addr)
	if err != nil {
		return err
	}
	srvr.logger.Info("starting REST server", slog.String("addr", srvr.Addr))
	go func() {
		if err = srvr.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
			srvr.logger.Error("failed to start REST server", slog.Any("error", err))
		}
	}()
	return nil
}

func (srvr *server) stop(ctx context.Context) error {
	srvr.logger.Info("stopping REST server", slog.String("addr", srvr.Addr))
	return srvr.Shutdown(ctx)
}
