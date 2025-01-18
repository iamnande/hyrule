package rest

import (
	"context"
	"encoding/json"
	"io"

	"github.com/iamnande/hyrule/internal/rest/apis/errors"
)

type Validateable interface {
	Validate() []errors.InvalidAttribute
}

// TODO: invalid params as span attributes
func BindJSON(ctx context.Context, in io.ReadCloser, obj Validateable) *errors.BaseError {
	if err := json.NewDecoder(in).Decode(obj); err != nil {
		return errors.NewInternalServerError(ctx, err)
	}
	if invalidAttributes := obj.Validate(); invalidAttributes != nil {
		return errors.NewBadRequestError(ctx, invalidAttributes...)
	}
	return nil
}
