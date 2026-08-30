package utils

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	iamJwksAPI "github.com/iamnande/hyrule/go/cmd/iam-jwks/app"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

type TestIamJwks struct {
	App *fxtest.App
	API http.Handler
}

func NewTestIamJwks(opts ...fx.Option) TestIamJwks {
	test := constructTestAPI([]fx.Option{fx.Provide(repository.LoadEnvConfig), iamJwksAPI.Module}, opts...)
	return TestIamJwks{
		API: test.API,
		App: test.App,
	}
}
