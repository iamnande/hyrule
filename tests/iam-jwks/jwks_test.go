package iamjwks

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/tests/utils"
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
		GinkgoT().Setenv("HYRULE_DATABASE_USER", "hyrule_app_ro")
		GinkgoT().Setenv("HYRULE_DATABASE_PASSWORD", "hyrule_app_ro")

		kid = fmt.Sprintf("test-%d", time.Now().UnixNano())

		pub, _, err := ed25519.GenerateKey(rand.Reader)
		Expect(err).ToNot(HaveOccurred())
		der, err := x509.MarshalPKIXPublicKey(pub)
		Expect(err).ToNot(HaveOccurred())
		pemKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
		wantX = base64.RawURLEncoding.EncodeToString(pub)

		conn, err := pgx.Connect(context.Background(), "postgres://hyrule_owner:hyrule_owner@localhost:5432/hyrule?sslmode=disable")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = conn.Close(context.Background()) }()
		_, err = conn.Exec(context.Background(),
			"INSERT INTO iam_jwks_keys (id, algorithm, public_key) VALUES ($1, $2, $3)",
			kid, "EdDSA", string(pemKey))
		Expect(err).ToNot(HaveOccurred())

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
