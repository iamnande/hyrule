package rest

import (
	"github.com/go-chi/chi/v5"
)

type APIHandler func(router chi.Router)
