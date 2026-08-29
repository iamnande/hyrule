package repository_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/iamnande/hyrule/internal/svc/iam-jwks/repository"
)

var _ = Describe("FileRepository", func() {
	Describe("List", func() {
		It("parses the keys file", func() {
			path := writeKeysFile(`[{"id":"kid-1","algorithm":"EdDSA","public_key":"pem","created_at":"2026-08-29T00:00:00Z"}]`)
			repo := repository.NewFile(repository.FileConfig{Path: path})

			keys, err := repo.List(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(keys).To(HaveLen(1))
			Expect(keys[0].ID).To(Equal("kid-1"))
			Expect(keys[0].Algorithm).To(Equal("EdDSA"))
			Expect(keys[0].PublicKey).To(Equal("pem"))
		})

		It("errors when the file doesn't exist", func() {
			repo := repository.NewFile(repository.FileConfig{Path: filepath.Join(GinkgoT().TempDir(), "missing.json")})

			_, err := repo.List(context.Background())

			Expect(err).To(HaveOccurred())
		})

		It("errors on malformed JSON", func() {
			path := writeKeysFile(`not json`)
			repo := repository.NewFile(repository.FileConfig{Path: path})

			_, err := repo.List(context.Background())

			Expect(err).To(HaveOccurred())
		})
	})
})

func writeKeysFile(contents string) string {
	path := filepath.Join(GinkgoT().TempDir(), "keys.json")
	Expect(os.WriteFile(path, []byte(contents), 0o600)).To(Succeed())
	return path
}
