package repository_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

var _ = Describe("EnvRepository", func() {
	Describe("List", func() {
		It("parses the keys", func() {
			repo := repository.NewEnv(repository.EnvConfig{
				Keys: `[{"id":"kid-1","algorithm":"EdDSA","public_key":"pem","created_at":"2026-08-29T00:00:00Z"}]`,
			})

			keys, err := repo.List(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(keys).To(HaveLen(1))
			Expect(keys[0].ID).To(Equal("kid-1"))
			Expect(keys[0].Algorithm).To(Equal("EdDSA"))
			Expect(keys[0].PublicKey).To(Equal("pem"))
		})

		It("errors on malformed JSON", func() {
			repo := repository.NewEnv(repository.EnvConfig{Keys: "not json"})

			_, err := repo.List(context.Background())

			Expect(err).To(HaveOccurred())
		})
	})
})
