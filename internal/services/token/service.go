package token

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

type Service struct {
	keyID        string
	verifyingKey *rsa.PublicKey
	signingKey   *rsa.PrivateKey
}

type Params struct {
	fx.In

	Config config.JWT
}

func NewService(params Params) (*Service, error) {
	privateKeyBytes := []byte(params.Config.PrivateKey)
	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKeyBytes := []byte(params.Config.PublicKey)
	keyID := fmt.Sprintf("%x", sha1.Sum(publicKeyBytes))
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &Service{
		keyID:        keyID,
		verifyingKey: publicKey,
		signingKey:   signingKey,
	}, nil
}

func (s *Service) GenerateToken(ctx context.Context) (Token, error) {
	trace := sentry.StartSpan(ctx, "service:token:GenerateToken")
	defer trace.Finish()
	generator := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{})
	token, err := generator.SignedString(s.signingKey)
	if err != nil {
		return Token{}, err
	}
	return Token{
		access:  token,
		refresh: "refresh",
		expires: time.Now(),
		eol:     time.Now(),
	}, nil
}
