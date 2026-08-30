package runtime

import (
	"log/slog"
	"net/http"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest"
)

type RouterParams struct {
	fx.In

	Logger    *slog.Logger
	HealthAPI rest.APIHandler   `name:"health"`
	APIs      []rest.APIHandler `group:"apis"`
}

// NewRouter is the router every service gets for free - a service
// contributes its own handlers via `group:"apis"`, nothing else. version
// lives in a header (see docs/conventions.md#url-structure--versioning),
// not the path, so there's no version segment here.
func NewRouter(params RouterParams) (http.Handler, error) {
	return rest.NewRouter(&rest.Config{
		Logger:    params.Logger,
		HealthAPI: params.HealthAPI,
		APIs:      params.APIs,
	}), nil
}
