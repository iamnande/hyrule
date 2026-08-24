package errors

import "net/http"

type ErrorCode string

func (code ErrorCode) String() string {
	return string(code)
}

type Definition struct {
	Code       ErrorCode
	Name       string
	Message    string
	StatusCode int
}

const (
	ErrorCodeInternal   ErrorCode = "ERR_HYRULE_000"
	ErrorCodeBadRequest ErrorCode = "ERR_HYRULE_100"
)

var registry = map[ErrorCode]Definition{
	ErrorCodeInternal: {
		Code:       ErrorCodeInternal,
		Name:       "Internal Server Error",
		Message:    "We seem to be having some trouble. Please notify support and provide the operation ID. We'll get right on it.",
		StatusCode: http.StatusInternalServerError,
	},
	ErrorCodeBadRequest: {
		Code:       ErrorCodeBadRequest,
		Name:       "Bad Request",
		Message:    "The request you made was invalid. Please check the input and try again.",
		StatusCode: http.StatusBadRequest,
	},
}
