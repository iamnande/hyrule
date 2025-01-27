package registration

import (
	"net/http"

	"github.com/iamnande/hyrule/internal/domains/registration"
	"github.com/iamnande/hyrule/internal/rest"
	apiErrors "github.com/iamnande/hyrule/internal/rest/apis/errors"
	"github.com/iamnande/hyrule/internal/rest/apis/response"
	"github.com/iamnande/hyrule/internal/rest/apis/validation"
)

type RegisterNewUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *RegisterNewUserRequest) Validate() []apiErrors.InvalidAttribute {
	var errors []apiErrors.InvalidAttribute
	if req.Email == "" {
		errors = append(errors, apiErrors.InvalidAttribute{
			Path:   "email",
			Rule:   validation.Required,
			Reason: "email is required",
			Source: apiErrors.AttributeSourceBody,
		})
	}
	if req.Password == "" {
		errors = append(errors, apiErrors.InvalidAttribute{
			Path:   "password",
			Rule:   validation.Required,
			Reason: "password is required",
			Source: apiErrors.AttributeSourceBody,
		})
	}
	if req.FullName == "" {
		errors = append(errors, apiErrors.InvalidAttribute{
			Path:   "full_name",
			Rule:   validation.Required,
			Reason: "full name is required",
			Source: apiErrors.AttributeSourceBody,
		})
	}
	return errors
}

func (api *API) RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		apiErr *apiErrors.BaseError

		ctx = r.Context()
		req = &RegisterNewUserRequest{}
	)

	if apiErr = rest.BindJSON(ctx, r.Body, req); apiErr != nil {
		response.JSONError(ctx, w, apiErr)
		return
	}

	_, err = api.registrationDomain.RegisterNewUser(ctx, &registration.RegisterNewUserInput{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	})
	if err != nil {
		response.JSONError(ctx, w, apiErrors.NewInternalServerError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
