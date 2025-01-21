package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"

	apiErrors "github.com/iamnande/hyrule/internal/rest/apis/errors"
	"github.com/iamnande/hyrule/internal/services/logging"
)

var (
	ErrFailedToSerializeResponse = errors.New("failed to serialize response")
	ErrFailedToWriteResponse     = errors.New("failed to write response")
)

type Response struct {
	Code int
	Data any
}

func JSON(_ context.Context, res http.ResponseWriter, code int, data any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	response, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Errorf("%w: %w", ErrFailedToSerializeResponse, err))
	}
	if _, err = res.Write(response); err != nil {
		panic(fmt.Errorf("%w: %w", ErrFailedToWriteResponse, err))
	}
}

func JSONError(ctx context.Context, res http.ResponseWriter, err *apiErrors.BaseError) {
	logger := logging.FromContext(ctx)
	logger.Error(err.Error())
	sentry.CaptureException(err)
	JSON(ctx, res, err.StatusCode, err)
}
