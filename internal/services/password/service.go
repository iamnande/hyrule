package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/getsentry/sentry-go"
	"go.uber.org/fx"
	"golang.org/x/crypto/argon2"

	"github.com/iamnande/hyrule/internal/config"
)

const (
	// EncodingFormat defines the storage format for password hashes:
	// $argon2id$v=<version>$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
	// where salt and hash are base64 encoded.
	EncodingFormat = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
)

var (
	ErrInvalidEncodingFormat = fmt.Errorf("invalid encoding format")
)

// Service implements password hashing using argon2.
// It includes a pepper for additional security and uses constant-time
// operations for comparing sensitive data.
type Service struct {
	memory     uint32
	iterations uint32
	threads    uint8
	keyLength  uint32
	saltLength uint32
	pepper     string
}

type Params struct {
	fx.In

	Config config.Password
}

func NewService(params Params) (*Service, error) {
	return &Service{
		memory:     params.Config.Memory,
		iterations: params.Config.Iterations,
		threads:    params.Config.Threads,
		keyLength:  params.Config.KeyLength,
		saltLength: params.Config.SaltLength,
		pepper:     params.Config.Pepper,
	}, nil
}

func (service *Service) salt() ([]byte, error) {
	salt := make([]byte, service.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func (service *Service) hash(password string, salt []byte) []byte {
	pepperedPassword := append([]byte(service.pepper), []byte(password)...)
	return argon2.IDKey(
		pepperedPassword,
		salt,
		service.iterations,
		service.memory,
		service.threads,
		service.keyLength,
	)
}

func (service *Service) encode(salt []byte, hash []byte) string {
	return fmt.Sprintf(EncodingFormat,
		argon2.Version,
		service.memory,
		service.iterations,
		service.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func (service *Service) decode(encoded string) ([]byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, ErrInvalidEncodingFormat
	}

	var version uint32
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, fmt.Errorf("%w: invalid version: %v", ErrInvalidEncodingFormat, err)
	}

	var memory uint32
	var iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
		return nil, nil, fmt.Errorf("%w: invalid parameters: %v", ErrInvalidEncodingFormat, err)
	}

	if version != argon2.Version ||
		memory != service.memory ||
		iterations != service.iterations ||
		threads != service.threads {
		return nil, nil, ErrInvalidEncodingFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid salt encoding: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid hash encoding: %w", err)
	}

	return salt, hash, nil
}

// HashPassword creates an argon2 hash of the password. The result is encoded
// in a standard format including the parameters used.
func (service *Service) HashPassword(ctx context.Context, password string) (string, error) {
	var (
		err error

		salt    []byte
		hash    []byte
		encoded string

		span = sentry.StartSpan(ctx, "services:password:HashPassword")
	)
	defer func() {
		if salt != nil {
			subtle.ConstantTimeCopy(1, salt, make([]byte, len(salt)))
		}
		if hash != nil {
			subtle.ConstantTimeCopy(1, hash, make([]byte, len(hash)))
		}
		span.Finish()
	}()

	salt, err = service.salt()
	if err != nil {
		return "", err
	}

	hash = service.hash(password, salt)
	encoded = service.encode(salt, hash)

	return encoded, nil
}

// VerifyPassword checks if a password matches its expected hash using
// constant-time comparison to prevent timing attacks.
func (service *Service) VerifyPassword(ctx context.Context, password string, expected string) (bool, error) {
	var (
		err error

		salt     []byte
		hash     []byte
		computed []byte

		span = sentry.StartSpan(ctx, "services:password:VerifyPassword")
	)
	defer func() {
		if salt != nil {
			subtle.ConstantTimeCopy(1, salt, make([]byte, len(salt)))
		}
		if hash != nil {
			subtle.ConstantTimeCopy(1, hash, make([]byte, len(hash)))
		}
		if computed != nil {
			subtle.ConstantTimeCopy(1, computed, make([]byte, len(computed)))
		}
		span.Finish()
	}()

	salt, hash, err = service.decode(expected)
	if err != nil {
		return false, err
	}

	computed = service.hash(password, salt)
	match := subtle.ConstantTimeCompare(hash, computed) == 1

	return match, nil
}
