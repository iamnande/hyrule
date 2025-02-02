package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/getsentry/sentry-go"

	"github.com/iamnande/hyrule/internal/rest/apis/errors"
)

type Validateable interface {
	Validate() errors.InvalidAttributes
}

func BindQuery(ctx context.Context, obj Validateable) *errors.BaseError {
	if invalidAttributes := obj.Validate(); invalidAttributes != nil {
		for _, invalidAttribute := range invalidAttributes {
			span := sentry.SpanFromContext(ctx)
			span.SetTag(fmt.Sprintf("request:attribute:%s:%s", invalidAttribute.Source, invalidAttribute.Path), invalidAttribute.String())
		}
		return errors.NewBadRequestError(ctx, invalidAttributes...)
	}
	return nil
}

func BindJSON(ctx context.Context, in io.ReadCloser, obj Validateable) *errors.BaseError {
	span := sentry.StartSpan(ctx, "rest:api:bind:json")
	defer span.Finish()
	if err := json.NewDecoder(in).Decode(obj); err != nil {
		return errors.NewInternalServerError(ctx, err)
	}
	if invalidAttributes := obj.Validate(); invalidAttributes != nil {
		for _, invalidAttribute := range invalidAttributes {
			span.SetTag(fmt.Sprintf("request:attribute:%s:%s", invalidAttribute.Source, invalidAttribute.Path), invalidAttribute.String())
		}
		return errors.NewBadRequestError(ctx, invalidAttributes...)
	}
	return nil
}
