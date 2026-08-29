package api

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	transporterrors "github.com/iamnande/hyrule/internal/lib/rest/transport/errors"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/domain"
)

type keyLister interface {
	List(ctx context.Context) ([]domain.Key, error)
}

type Handlers struct {
	service keyLister
}

func NewHandlers(service keyLister) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) GetJWKS(ctx context.Context, _ GetJWKSRequestObject) (GetJWKSResponseObject, error) {
	keys, err := h.service.List(ctx)
	if err != nil {
		return GetJWKS500JSONResponse{toInternalError(transporterrors.NewInternalServerError(ctx, err))}, nil
	}
	jwks := make([]JWK, len(keys))
	for i, key := range keys {
		jwk, err := toJWK(key)
		if err != nil {
			return GetJWKS500JSONResponse{toInternalError(transporterrors.NewInternalServerError(ctx, err))}, nil
		}
		jwks[i] = jwk
	}
	return GetJWKS200JSONResponse{Keys: jwks}, nil
}

func toJWK(key domain.Key) (JWK, error) {
	block, _ := pem.Decode([]byte(key.PublicKey))
	if block == nil {
		return JWK{}, fmt.Errorf("key %q: no PEM block found", key.ID)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return JWK{}, fmt.Errorf("key %q: parse public key: %w", key.ID, err)
	}
	edKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return JWK{}, fmt.Errorf("key %q: not an Ed25519 public key", key.ID)
	}
	return JWK{
		Kid: key.ID,
		Kty: OKP,
		Alg: JWKAlg(key.Algorithm),
		Use: Sig,
		Crv: Ed25519,
		X:   base64.RawURLEncoding.EncodeToString(edKey),
	}, nil
}

func toInternalError(err *transporterrors.BaseError) InternalServerErrorJSONResponse {
	return InternalServerErrorJSONResponse{
		Code:        err.Code.String(),
		Name:        err.Name,
		Message:     err.Message,
		OperationId: err.OperationID,
	}
}
