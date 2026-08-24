package rest

import (
	"github.com/go-chi/chi/v5"
)

// APIHandler registers its routes directly on the shared router. every
// service's paths are flat and absolute (see
// docs/conventions.md#url-structure--versioning), so handlers register
// side by side on one router instead of each claiming their own Mount -
// chi only allows one Mount per pattern, and root-level paths would all
// collide on "/".
type APIHandler func(router chi.Router)
