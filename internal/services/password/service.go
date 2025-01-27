package password

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/getsentry/sentry-go"
	"go.uber.org/fx"
	"golang.org/x/crypto/argon2"
)

const (
	EncodingFormat = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
)

type Service struct {
	memory     uint32
	iterations uint32
	threads    uint8
	keyLength  uint32
	saltLength uint32
}

type Params struct {
	fx.In
}

func NewService(params Params) (*Service, error) {
	return &Service{
		memory:     1024 * 64,
		iterations: 3,
		threads:    4,
		keyLength:  32,
		saltLength: 16,
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
	return argon2.IDKey(
		[]byte(password),
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
	var (
		err        error
		version    uint32
		memory     uint32
		iterations uint32
		threads    uint8
	)
	if _, err = fmt.Sscanf(encoded, EncodingFormat, &version, &memory, &iterations, &threads); err != nil {
		return nil, nil, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, nil, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, nil, err
	}
	return salt, hash, nil
}

func (service *Service) HashPassword(ctx context.Context, password string) (string, error) {
	var (
		err error

		span = sentry.StartSpan(ctx, "services:password:HashPassword")
	)
	defer span.Finish()

	var salt []byte
	salt, err = service.salt()
	if err != nil {
		return "", err
	}

	hash := service.hash(password, salt)
	encoded := service.encode(salt, hash)

	return encoded, nil
}

func (service *Service) VerifyPassword(ctx context.Context, password string, expected string) (bool, error) {
	var (
		err error

		span = sentry.StartSpan(ctx, "services:password:VerifyPassword")
	)
	defer span.Finish()

	salt, hashed, err := service.decode(expected)
	if err != nil {
		return false, err
	}

	actual := service.hash(password, salt)

	if match := bytes.Equal(hashed, actual); match {
		return true, nil
	}

	return false, nil
}
