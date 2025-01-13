package rest

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"go.uber.org/fx"
)

type server struct {
	*http.Server

	logger *slog.Logger
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger *slog.Logger
	API    http.Handler
}

func NewServer(params Params) *http.Server {
	srvr := &server{
		logger: params.Logger,
		Server: &http.Server{
			Addr:    ":8000", // maybe configurable, maybe not
			Handler: params.API,
			// TODO: sane default timeouts
			// TODO: configurable timeouts
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
