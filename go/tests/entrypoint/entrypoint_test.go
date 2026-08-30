package entrypoint

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	iamJwksAPI "github.com/iamnande/hyrule/go/cmd/iam-jwks/app"
	pingsAPI "github.com/iamnande/hyrule/go/cmd/pings/app"
	"github.com/iamnande/hyrule/go/internal/lib/runtime"
	"github.com/iamnande/hyrule/go/tests/utils"
)

func TestEntrypoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Entrypoint Suite")
}

var _ = DescribeTable("Entrypoint",
	func(opts []fx.Option, invoke fx.Option) {
		initialized := false

		opts = append(opts, utils.TestConfigs()...)
		opts = append(opts, fx.Invoke(func() { initialized = true }))
		opts = append(opts, invoke)

		fxtest.New(GinkgoT(), opts...).
			RequireStart().
			RequireStop()

		Expect(initialized).To(BeTrue())
	},
	Entry(
		"Pings",
		[]fx.Option{runtime.NewModule(pingsAPI.Name), pingsAPI.Module},
		fx.Invoke(func(app fx.Shutdowner) { _ = app.Shutdown() }),
	),
	Entry(
		"IamJwks",
		[]fx.Option{runtime.NewModule(iamJwksAPI.Name), iamJwksAPI.Module},
		fx.Invoke(func(app fx.Shutdowner) { _ = app.Shutdown() }),
	),
)
