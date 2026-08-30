package pings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/imroc/req/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/tests/utils"
)

var _ = Describe("PingsAPI", Ordered, ContinueOnFailure, func() {
	var (
		baseURL       string
		testAPIServer *httptest.Server
		pingsAPI      utils.TestPings
		name          string
	)

	BeforeAll(func() {
		opts := []fx.Option{}
		opts = append(opts, utils.TestConfigs()...)
		pingsAPI = utils.NewTestPings(opts...)
		testAPIServer = httptest.NewServer(pingsAPI.API)
		baseURL = testAPIServer.URL
		name = fmt.Sprintf("test-%d", time.Now().UnixNano())
	})

	AfterAll(func() {
		testAPIServer.Close()
		pingsAPI.App.RequireStop()
	})

	Describe("POST /pings", func() {
		Context("with a valid name and kind", func() {
			It("registers it and returns it as up", func() {
				call, err := req.R().SetBody(map[string]string{"name": name, "kind": "host"}).Post(baseURL + "/pings")
				Expect(err).ToNot(HaveOccurred())
				Expect(call.Response).To(HaveHTTPStatus(http.StatusOK))
				Expect(call.String()).To(SatisfyAll(
					ContainSubstring(fmt.Sprintf(`"name":"%s"`, name)),
					ContainSubstring(`"kind":"host"`),
					ContainSubstring(`"state":"up"`),
				))
			})
		})

		Context("with an invalid kind", func() {
			It("rejects it with a bad request naming the bad attribute", func() {
				call, err := req.R().SetBody(map[string]string{"name": "whatever", "kind": "toaster"}).Post(baseURL + "/pings")
				Expect(err).ToNot(HaveOccurred())
				Expect(call.Response).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(call.String()).To(ContainSubstring(`"path":"kind"`))
			})
		})

		Context("with a missing name", func() {
			It("rejects it with a bad request naming the bad attribute", func() {
				call, err := req.R().SetBody(map[string]string{"name": "", "kind": "host"}).Post(baseURL + "/pings")
				Expect(err).ToNot(HaveOccurred())
				Expect(call.Response).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(call.String()).To(ContainSubstring(`"path":"name"`))
			})
		})
	})

	Describe("GET /pings", func() {
		Context("after a name has pinged in", func() {
			It("lists it", func() {
				call, err := req.Get(baseURL + "/pings")
				Expect(err).ToNot(HaveOccurred())
				Expect(call.Response).To(HaveHTTPStatus(http.StatusOK))
				Expect(call.String()).To(ContainSubstring(fmt.Sprintf(`"name":"%s"`, name)))
			})
		})
	})
})
