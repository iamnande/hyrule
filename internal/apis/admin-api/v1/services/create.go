package services

import (
	"net/http"

	"github.com/iamnande/hyrule/internal/domains/admin"
)

type CreateRequest struct {
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (request *CreateRequest) Validate() error {
	// TODO: should live in rest. must return all errors at once
	return nil
}

func (api *API) CreateService(w http.ResponseWriter, r *http.Request) {
	// TODO: wire up tracing
	// TODO: rest - parse json
	// TODO: rest - validation
	service, err := api.adminDomain.CreateService(r.Context(), &admin.CreateServiceRequest{
		Domain:      "platform",
		Name:        "admin-api",
		Description: "faker.Lorem().Sentence(10)",
	})
	if err != nil {
		// TODO: rest - status code error response
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	// TODO: rest - marshal json
	_, _ = w.Write([]byte(service.ID.String()))
}
