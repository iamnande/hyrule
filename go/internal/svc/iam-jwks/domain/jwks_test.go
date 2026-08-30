package domain_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/domain"
)

type fakeKeyStore struct {
	keys []domain.Key
	err  error
}

func (f *fakeKeyStore) List(ctx context.Context) ([]domain.Key, error) {
	return f.keys, f.err
}

var _ = Describe("Service", func() {
	Describe("List", func() {
		It("returns the keys from the store", func() {
			want := []domain.Key{{ID: "kid-1", Algorithm: "RS256", PublicKey: "pem"}}
			svc := domain.NewService(&fakeKeyStore{keys: want})

			got, err := svc.List(context.Background())

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		})

		It("propagates a store error", func() {
			svc := domain.NewService(&fakeKeyStore{err: errors.New("boom")})

			_, err := svc.List(context.Background())

			Expect(err).To(MatchError("boom"))
		})
	})
})
