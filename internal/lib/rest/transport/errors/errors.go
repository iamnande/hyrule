package errors

import (
	"context"
	"fmt"
	"strings"

	"github.com/iamnande/hyrule/internal/lib/rest/transport/validation"
	"github.com/iamnande/hyrule/internal/lib/tracing"
	"github.com/iamnande/hyrule/internal/lib/version"
)

type AttributeSource string

const (
	AttributeSourceBody   AttributeSource = "body"
	AttributeSourcePath   AttributeSource = "path"
	AttributeSourceQuery  AttributeSource = "query"
	AttributeSourceHeader AttributeSource = "header"
)

type BaseError struct {
	Code              ErrorCode          `json:"code"`
	Name              string             `json:"name"`
	Message           string             `json:"message"`
	OperationID       string             `json:"operation_id"`
	InvalidAttributes []InvalidAttribute `json:"invalid_attributes,omitempty"`

	// not marshalled, used for response/logging
	StatusCode    int   `json:"-"`
	InternalError error `json:"-"`
}

func (err *BaseError) Error() string {
	if len(err.InvalidAttributes) > 0 {
		return fmt.Sprintf("[%s] %+v (%s)",
			strings.ToUpper(err.Code.String()),
			err.InvalidAttributes,
			err.OperationID,
		)
	}
	return fmt.Sprintf("[%s] %s - %v (%s)",
		strings.ToUpper(err.Code.String()),
		err.Name,
		err.InternalError,
		err.OperationID,
	)
}

func (err *BaseError) Unwrap() error {
	return err.InternalError
}

type InvalidAttributes []InvalidAttribute

func (invalidAttributes InvalidAttributes) String() string {
	var (
		invalidAttributeStrings []string
	)
	for _, invalidAttribute := range invalidAttributes {
		invalidAttributeStrings = append(invalidAttributeStrings, invalidAttribute.String())
	}
	return strings.Join(invalidAttributeStrings, ", ")
}

type InvalidAttribute struct {
	Path   string          `json:"path"`
	Rule   validation.Rule `json:"rule"`
	Reason string          `json:"reason"`
	Source AttributeSource `json:"source"`
}

func (invalidAttribute InvalidAttribute) String() string {
	return fmt.Sprintf(
		"'%s' in %s failed validation: %s (%s)",
		invalidAttribute.Path,
		invalidAttribute.Source,
		invalidAttribute.Reason,
		invalidAttribute.Rule,
	)
}

// newError looks up code's Definition in the registry and builds the
// per-request shell around it - constructors below only attach what's
// specific to this call (the wrapped error, the invalid attributes).
func newError(ctx context.Context, code ErrorCode) *BaseError {
	def := registry[code]
	return &BaseError{
		Code:        def.Code,
		Name:        def.Name,
		Message:     def.Message,
		StatusCode:  def.StatusCode,
		OperationID: extractTraceIDFromContext(ctx),
	}
}

func NewInternalServerError(ctx context.Context, err error) *BaseError {
	base := newError(ctx, ErrorCodeInternal)
	base.InternalError = err
	return base
}

func NewBadRequestError(ctx context.Context, invalidAttributes ...InvalidAttribute) *BaseError {
	base := newError(ctx, ErrorCodeBadRequest)
	base.InvalidAttributes = invalidAttributes
	return base
}

func extractTraceIDFromContext(ctx context.Context) string {
	traceID, ok := tracing.TraceID(ctx)
	if !ok {
		return fmt.Sprintf("%s:trace:unknown", version.ServicePrefix)
	}
	return fmt.Sprintf("%s:trace:%s", version.ServicePrefix, traceID)
}
