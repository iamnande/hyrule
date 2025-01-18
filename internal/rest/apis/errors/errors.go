package errors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

type InvalidAttribute struct {
	Path   string          `json:"path"`
	Rule   validation.Rule `json:"rule"`
	Reason string          `json:"reason"`
	Source AttributeSource `json:"source"`
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
	// NOTE: this is a placeholder for now
	return "hyrule:trace:trace-id"
}
