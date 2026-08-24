package errors

import "net/http"

// ErrorCode is a stable, catalog-scoped identifier - ERR_HYRULE_NNN,
// grouped by numeric range per domain. the shape mirrors ngrok's error
// reference (https://ngrok.com/docs/errors/reference): a code is meant to
// be stable enough to depend on and to eventually resolve to a docs page,
// one code per class of failure. the codegen-from-config half of that
// (and the docs page) is future work - this is the record-keeping start.
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
	// 000-099: internal/system - not the caller's fault.
	ErrorCodeInternal ErrorCode = "ERR_HYRULE_000"

	// 100-199: request validation - the caller's fault.
	ErrorCodeBadRequest ErrorCode = "ERR_HYRULE_100"
)

// registry is the single source of truth for every defined error code -
// hand-written today, but shaped like what a generator would emit from a
// config file, so swapping the source later is mechanical rather than a
// redesign.
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
