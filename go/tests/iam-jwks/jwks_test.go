package iamjwks

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
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

var _ = Describe("JWKS API", Ordered, ContinueOnFailure, func() {
	var (
		baseURL       string
		testAPIServer *httptest.Server
		jwksAPI       utils.TestIamJwks
		kid           string
		wantX         string
	)

	BeforeAll(func() {
		kid = fmt.Sprintf("test-%d", time.Now().UnixNano())

		pub, _, err := ed25519.GenerateKey(rand.Reader)
		Expect(err).ToNot(HaveOccurred())
		der, err := x509.MarshalPKIXPublicKey(pub)
		Expect(err).ToNot(HaveOccurred())
		pemKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
		wantX = base64.RawURLEncoding.EncodeToString(pub)

		keys, err := json.Marshal([]map[string]string{{
			"id":         kid,
			"algorithm":  "EdDSA",
			"public_key": string(pemKey),
			"created_at": time.Now().UTC().Format(time.RFC3339),
		}})
		Expect(err).ToNot(HaveOccurred())
		GinkgoT().Setenv("HYRULE_IAM_JWKS_KEYS", string(keys))

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

	Describe("GET /.well-known/jwks.json", func() {
		It("serves the seeded key as a JWK", func() {
			call, err := req.Get(baseURL + "/.well-known/jwks.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(call.Response).To(HaveHTTPStatus(http.StatusOK))
			Expect(call.String()).To(SatisfyAll(
				ContainSubstring(fmt.Sprintf(`"kid":"%s"`, kid)),
				ContainSubstring(`"kty":"OKP"`),
				ContainSubstring(`"alg":"EdDSA"`),
				ContainSubstring(`"use":"sig"`),
				ContainSubstring(`"crv":"Ed25519"`),
				ContainSubstring(fmt.Sprintf(`"x":"%s"`, wantX)),
			))
		})
	})
})
