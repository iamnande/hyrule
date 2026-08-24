package utils

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	modules "github.com/iamnande/hyrule/internal/lib/runtime"
	"github.com/iamnande/hyrule/internal/lib/version"
)

type TestAPI struct {
	fx.In

	App *fxtest.App
	API http.Handler
}

func constructTestAPI(appOptions []fx.Option, opts ...fx.Option) TestAPI {
	type Target struct {
		fx.In

		H http.Handler
	}
	var t Target

	options := []fx.Option{}
	options = append(options, fx.Supply(version.NewServiceInfo("test")))
	options = append(options, modules.LoggingModule)
	options = append(options, modules.HTTPModule)
	options = append(options, appOptions...)
	options = append(options, opts...)
	options = append(options, fx.Populate(&t))

	app := fxtest.New(GinkgoT(), options...)
	app.RequireStart()

	if t.H == nil {
		panic("no API handler was provided")
	}

	return TestAPI{
		API: t.H,
		App: app,
	}
}
