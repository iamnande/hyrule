package utils

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	registrationApp "github.com/iamnande/hyrule/cmd/registration-api/app"
)

type TestRegistrationAPI struct {
	App *fxtest.App
	API http.Handler
}

func NewTestRegistrationAPI(opts ...fx.Option) TestRegistrationAPI {
	test := constructTestAPI(registrationApp.Build(), opts...)
	return TestRegistrationAPI{
		API: test.API,
		App: test.App,
	}
}
