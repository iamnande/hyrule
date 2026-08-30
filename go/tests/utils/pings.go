package utils

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	pingsAPI "github.com/iamnande/hyrule/go/cmd/pings/app"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
)

type TestPings struct {
	App *fxtest.App
	API http.Handler
}

func NewTestPings(opts ...fx.Option) TestPings {
	test := constructTestAPI([]fx.Option{pingsAPI.Module, fx.Provide(domain.LoadConfig)}, opts...)
	return TestPings{
		API: test.API,
		App: test.App,
	}
}
