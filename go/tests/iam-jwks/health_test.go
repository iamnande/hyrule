package iamjwks

import (
	"net/http"
	"net/http/httptest"

	"github.com/imroc/req/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/tests/utils"
)

var _ = Describe("HealthAPI", Ordered, ContinueOnFailure, func() {
	var (
		baseURL       string
		testAPIServer *httptest.Server
		jwksAPI       utils.TestIamJwks
	)

	BeforeAll(func() {
		opts := []fx.Option{}
		opts = append(opts, utils.TestConfigs()...)
		jwksAPI = utils.NewTestIamJwks(opts...)
		testAPIServer = httptest.NewServer(jwksAPI.API)
		baseURL = testAPIServer.URL
	})

	AfterAll(func() {
		testAPIServer.Close()
		jwksAPI.App.RequireStop()
	})

	DescribeTable("when we call the HealthAPI route",
		func(route string) {
			call, err := req.Get(baseURL + route)
			Expect(err).ToNot(HaveOccurred())
			Expect(call.Response).To(HaveHTTPStatus(http.StatusOK))
		},
		Entry("/discovery, we receive an HTTP 200 OK success response back", "/discovery"),
		Entry("/healthz, we receive an HTTP 200 OK success response back", "/healthz"),
		Entry("/startupz, we receive an HTTP 200 OK success response back", "/startupz"),
		Entry("/livez, we receive an HTTP 200 OK success response back", "/livez"),
		Entry("/readyz, we receive an HTTP 200 OK success response back", "/readyz"),
	)

	Describe("the /healthz route", func() {
		It("reports itself up with no dependencies - iam-jwks doesn't touch postgres anymore", func() {
			call, err := req.Get(baseURL + "/healthz")
			Expect(err).ToNot(HaveOccurred())
			Expect(call.String()).To(SatisfyAll(
				ContainSubstring(`"status":"up"`),
				ContainSubstring(`"dependencies":[]`),
			))
		})
	})
})
