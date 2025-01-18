package signup

import (
	"net/http"

	"github.com/iamnande/hyrule/internal/rest"
	apiErrors "github.com/iamnande/hyrule/internal/rest/apis/errors"
	"github.com/iamnande/hyrule/internal/rest/apis/response"
	"github.com/iamnande/hyrule/internal/rest/apis/validation"
	"github.com/iamnande/hyrule/internal/services/signup"
)

type SignUpRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (request *SignUpRequest) Validate() []apiErrors.InvalidAttribute {
	invalidAttributes := make([]apiErrors.InvalidAttribute, 0)
	if request.FirstName == "" {
		invalidAttributes = append(invalidAttributes, apiErrors.InvalidAttribute{
			Path:   "first_name",
			Rule:   validation.Required,
			Reason: "first name is required", // TODO: path + is required (rule)
			Source: apiErrors.AttributeSourceBody,
		})
	}
	if request.LastName == "" {
		invalidAttributes = append(invalidAttributes, apiErrors.InvalidAttribute{
			Path:   "last_name",
			Rule:   validation.Required,
			Reason: "last name is required", // TODO: path + is required (rule)
			Source: apiErrors.AttributeSourceBody,
		})
	}
	if request.Email == "" {
		invalidAttributes = append(invalidAttributes, apiErrors.InvalidAttribute{
			Path:   "email",
			Rule:   validation.Required,
			Reason: "email is required", // TODO: path + is required (rule)
			Source: apiErrors.AttributeSourceBody,
		})
	}
	if request.Password == "" {
		invalidAttributes = append(invalidAttributes, apiErrors.InvalidAttribute{
			Path:   "password",
			Rule:   validation.Required,
			Reason: "password is required", // TODO: path + is required (rule)
			Source: apiErrors.AttributeSourceBody,
		})
	}
	if len(invalidAttributes) == 0 {
		return nil
	}
	return invalidAttributes
}

func (api *API) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := SignUpRequest{}
	if err := rest.BindJSON(ctx, r.Body, &req); err != nil {
		response.JSONError(ctx, w, err)
		return
	}

	// TODO: wire up tracing
	user, err := api.signUpService.Signup(ctx, &signup.SignUpRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		response.JSONError(ctx, w, apiErrors.NewInternalServerError(ctx, err))
		return
	}

	response.JSON(ctx, w, http.StatusOK, user)
}
