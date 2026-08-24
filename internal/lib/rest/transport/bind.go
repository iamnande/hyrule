package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/iamnande/hyrule/internal/lib/rest/transport/errors"
	"github.com/iamnande/hyrule/internal/lib/tracing"
)

type Validateable interface {
	Validate() errors.InvalidAttributes
}

func BindQuery(ctx context.Context, obj Validateable) *errors.BaseError {
	if invalidAttributes := obj.Validate(); invalidAttributes != nil {
		for _, invalidAttribute := range invalidAttributes {
			tracing.SetTag(ctx, fmt.Sprintf("request.attribute.%s.%s", invalidAttribute.Source, invalidAttribute.Path), invalidAttribute.String())
		}
		return errors.NewBadRequestError(ctx, invalidAttributes...)
	}
	return nil
}

func BindJSON(ctx context.Context, in io.ReadCloser, obj Validateable) *errors.BaseError {
	ctx, done := tracing.Start(ctx)
	defer done()
	if err := json.NewDecoder(in).Decode(obj); err != nil {
		return errors.NewInternalServerError(ctx, err)
	}
	if invalidAttributes := obj.Validate(); invalidAttributes != nil {
		for _, invalidAttribute := range invalidAttributes {
			tracing.SetTag(ctx, fmt.Sprintf("request.attribute.%s.%s", invalidAttribute.Source, invalidAttribute.Path), invalidAttribute.String())
		}
		return errors.NewBadRequestError(ctx, invalidAttributes...)
	}
	return nil
}
