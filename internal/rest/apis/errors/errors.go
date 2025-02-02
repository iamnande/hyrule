package errors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"

	"github.com/iamnande/hyrule/internal/rest/apis/validation"
)

type ErrorCode string

const (
	ErrorCodeInternal   ErrorCode = "internal"
	ErrorCodeBadRequest ErrorCode = "bad-request"
)

func (code ErrorCode) String() string {
	return fmt.Sprintf("api.hyhrule.local/errors/%s", string(code))
}

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

func NewInternalServerError(ctx context.Context, err error) *BaseError {
	return &BaseError{
		Code:          ErrorCodeInternal,
		Name:          "Internal Server Error",
		Message:       "We seem to be having some trouble. Please notify support and provide the operation ID. We'll get right on it.",
		OperationID:   extractTraceIDFromContext(ctx),
		StatusCode:    http.StatusInternalServerError,
		InternalError: err,
	}
}

func NewBadRequestError(ctx context.Context, invalidAttributes ...InvalidAttribute) *BaseError {
	return &BaseError{
		Code:              ErrorCodeBadRequest,
		Name:              "Bad Request",
		Message:           "The request you made was invalid. Please check the input and try again.",
		OperationID:       extractTraceIDFromContext(ctx),
		InvalidAttributes: invalidAttributes,
		StatusCode:        http.StatusBadRequest,
	}
}

func extractTraceIDFromContext(ctx context.Context) string {
	span := sentry.SpanFromContext(ctx)
	return fmt.Sprintf("hyrule:trace:%s", span.TraceID.String())
}
