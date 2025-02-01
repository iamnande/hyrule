package password

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/argon2"

	"github.com/iamnande/hyrule/internal/config"
)

var _ = Describe("Password Service", func() {
	var (
		ctx     context.Context
		service *Service
	)

	BeforeEach(func() {
		var (
			err error
		)
		ctx = context.Background()

		service, err = NewService(Params{
			Config: config.Password{
				Memory:     64 * 1024,
				Iterations: 3,
				Threads:    4,
				KeyLength:  32,
				SaltLength: 16,
				Pepper:     "test-pepper",
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("HashPassword", func() {
		It("should generate different hashes for the same password", func() {
			hash1, err := service.HashPassword(ctx, "password123")
			Expect(err).NotTo(HaveOccurred())

			hash2, err := service.HashPassword(ctx, "password123")
			Expect(err).NotTo(HaveOccurred())

			Expect(hash1).NotTo(Equal(hash2))
		})
		It("should generate a hash in the correct format", func() {
			hash, err := service.HashPassword(ctx, "password123")
			Expect(err).NotTo(HaveOccurred())

			parts := strings.Split(hash, "$")
			Expect(parts).To(HaveLen(6))
			Expect(parts[1]).To(Equal("argon2id"))

			var version uint32
			var memory uint32
			var iterations uint32
			var threads uint8
			_, err = fmt.Sscanf(parts[2], "v=%d", &version)
			Expect(err).NotTo(HaveOccurred())
			Expect(int(version)).To(Equal(argon2.Version))

			_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads)
			Expect(err).NotTo(HaveOccurred())
			Expect(memory).To(Equal(service.memory))
			Expect(iterations).To(Equal(service.iterations))
			Expect(threads).To(Equal(service.threads))

			// Validate salt and hash are base64 encoded
			_, err = base64.RawStdEncoding.DecodeString(parts[4])
			Expect(err).NotTo(HaveOccurred())
			_, err = base64.RawStdEncoding.DecodeString(parts[5])
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("VerifyPassword", func() {
		It("should verify a correct password", func() {
			password := "password123"
			hash, err := service.HashPassword(ctx, password)
			Expect(err).NotTo(HaveOccurred())

			match, err := service.VerifyPassword(ctx, password, hash)
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeTrue())
		})

		It("should reject an incorrect password", func() {
			hash, err := service.HashPassword(ctx, "password123")
			Expect(err).NotTo(HaveOccurred())

			match, err := service.VerifyPassword(ctx, "wrongpassword", hash)
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		It("should reject a malformed hash", func() {
			match, err := service.VerifyPassword(ctx, "password123", "malformed-hash")
			Expect(err).To(MatchError(ErrInvalidEncodingFormat))
			Expect(match).To(BeFalse())
		})

		It("should reject a hash with invalid parameters", func() {
			// Create a hash with different parameters
			differentService, err := NewService(Params{
				Config: config.Password{
					Memory:     32 * 1024, // Different memory value
					Iterations: 3,
					Threads:    4,
					KeyLength:  32,
					SaltLength: 16,
					Pepper:     "test-pepper",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			hash, err := differentService.HashPassword(ctx, "password123")
			Expect(err).NotTo(HaveOccurred())

			match, err := service.VerifyPassword(ctx, "password123", hash)
			Expect(err).To(MatchError(ErrInvalidEncodingFormat))
			Expect(match).To(BeFalse())
		})

		It("should reject a hash with invalid base64 salt", func() {
			hash := fmt.Sprintf(EncodingFormat,
				argon2.Version,
				service.memory,
				service.iterations,
				service.threads,
				"invalid-base64",
				"validbase64")
			match, err := service.VerifyPassword(ctx, "password123", hash)
			Expect(err).To(HaveOccurred())
			Expect(match).To(BeFalse())
		})

		It("should reject a hash with invalid base64 hash", func() {
			salt := make([]byte, service.saltLength)
			hash := fmt.Sprintf(EncodingFormat,
				argon2.Version,
				service.memory,
				service.iterations,
				service.threads,
				base64.RawStdEncoding.EncodeToString(salt),
				"invalid-base64")
			match, err := service.VerifyPassword(ctx, "password123", hash)
			Expect(err).To(HaveOccurred())
			Expect(match).To(BeFalse())
		})
	})
})
