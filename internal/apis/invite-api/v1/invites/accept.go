package invites

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/segmentio/ksuid"

	"github.com/iamnande/hyrule/internal/rest"
	apiErrors "github.com/iamnande/hyrule/internal/rest/apis/errors"
	"github.com/iamnande/hyrule/internal/rest/apis/response"
	"github.com/iamnande/hyrule/internal/rest/apis/validation"
)

type AcceptInviteRequest struct {
	Token string
}

func (req *AcceptInviteRequest) Validate() apiErrors.InvalidAttributes {
	var errors apiErrors.InvalidAttributes
	if req.Token == "" {
		errors = append(errors, apiErrors.InvalidAttribute{
			Path:   "token",
			Rule:   validation.Required,
			Reason: "token is required",
			Source: apiErrors.AttributeSourcePath,
		})
	}
	if _, err := ksuid.Parse(req.Token); err != nil {
		errors = append(errors, apiErrors.InvalidAttribute{
			Path:   "token",
			Rule:   validation.Unsupported,
			Reason: "token is invalid",
			Source: apiErrors.AttributeSourcePath,
		})
	}
	return errors
}

func (api *API) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var (
		apiErr *apiErrors.BaseError

		ctx = r.Context()
		req = &AcceptInviteRequest{
			Token: chi.URLParam(r, "token"),
		}
	)

	if apiErr = rest.BindQuery(ctx, req); apiErr != nil {
		response.JSONError(ctx, w, apiErr)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
