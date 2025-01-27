package adminapi

import (
	"net/http"
	"net/http/httptest"

	"github.com/imroc/req/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/tests/utils"
)

var _ = Describe("Health API", Ordered, ContinueOnFailure, func() {
	var (
		baseURL       string
		testAPIServer *httptest.Server
		signup        utils.TestRegistrationAPI
	)

	BeforeAll(func() {
		opts := []fx.Option{}
		opts = append(opts, utils.TestConfigs()...)
		signup = utils.NewTestRegistrationAPI(opts...)
		testAPIServer = httptest.NewServer(signup.API)
		baseURL = testAPIServer.URL
	})

	AfterAll(func() {
		testAPIServer.Close()
		signup.App.RequireStop()
	})

	DescribeTable("Health API",
		func(urlPath string) {
			uri := baseURL + urlPath
			call, err := req.Get(uri)
			Expect(err).ToNot(HaveOccurred())
			Expect(call.Response).To(HaveHTTPStatus(http.StatusOK))
		},
		Entry("/discovery", "/discovery"),
		Entry("/health/dependencies", "/health/dependencies"),
		Entry("/health/probes/liveness", "/health/probes/liveness"),
		Entry("/health/probes/readiness", "/health/probes/readiness"),
	)
})
