package pings

import (
	"net/http"
	"net/http/httptest"

	"github.com/imroc/req/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/tests/utils"
)

var _ = Describe("HealthAPI", Ordered, ContinueOnFailure, func() {
	var (
		baseURL       string
		testAPIServer *httptest.Server
		pingsAPI      utils.TestPings
	)

	BeforeAll(func() {
		opts := []fx.Option{}
		opts = append(opts, utils.TestConfigs()...)
		pingsAPI = utils.NewTestPings(opts...)
		testAPIServer = httptest.NewServer(pingsAPI.API)
		baseURL = testAPIServer.URL
	})

	AfterAll(func() {
		testAPIServer.Close()
		pingsAPI.App.RequireStop()
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

	Describe("the /discovery route", func() {
		Context("when we call it", func() {
			It("responds with this service's own metadata", func() {
				call, err := req.Get(baseURL + "/discovery")
				Expect(err).ToNot(HaveOccurred())
				Expect(call.String()).To(SatisfyAll(
					ContainSubstring(`"name":"hyrule-test"`),
					ContainSubstring(`"environment":"local"`),
					ContainSubstring(`"region":"us-east-2"`),
				))
			})
		})
	})

	Describe("the /healthz route", func() {
		Context("with the database reachable", func() {
			It("reports itself up with the database listed as a hard dependency", func() {
				call, err := req.Get(baseURL + "/healthz")
				Expect(err).ToNot(HaveOccurred())
				Expect(call.String()).To(SatisfyAll(
					ContainSubstring(`"status":"up"`),
					ContainSubstring(`"name":"database"`),
					ContainSubstring(`"status":"up"`),
					ContainSubstring(`"type":"hard"`),
				))
			})
		})
	})
})
