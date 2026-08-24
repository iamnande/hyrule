package rest

import (
	"net/http"
)

type APIHandler interface {
	Handler() http.Handler
	URLPath() string
}
