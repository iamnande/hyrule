package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/iamnande/hyrule/internal/lib/rest"
)

func NewRouter(handlers *Handlers) rest.APIHandler {
	strict := NewStrictHandler(handlers, nil)
	return func(router chi.Router) {
		HandlerFromMux(strict, router)
	}
}
