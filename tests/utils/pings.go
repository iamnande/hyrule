package utils

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	pingsAPI "github.com/iamnande/hyrule/cmd/pings/app"
)

type TestPings struct {
	App *fxtest.App
	API http.Handler
}

func NewTestPings(opts ...fx.Option) TestPings {
	test := constructTestAPI([]fx.Option{pingsAPI.Module}, opts...)
	return TestPings{
		API: test.API,
		App: test.App,
	}
}
