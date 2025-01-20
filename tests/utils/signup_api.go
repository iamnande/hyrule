package utils

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	signupApp "github.com/iamnande/hyrule/cmd/signup-api/app"
)

type TestSignUpAPI struct {
	App *fxtest.App
	API http.Handler
}

func NewTestSignUpAPI(opts ...fx.Option) TestSignUpAPI {
	test := constructTestAPI(signupApp.Build(), opts...)
	return TestSignUpAPI{
		API: test.API,
		App: test.App,
	}
}
