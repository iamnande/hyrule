package user

import (
	"go.uber.org/fx"
)

type Repository struct {
	// TODO: db/repo config
	// TODO: db client interface
}

type Params struct {
	fx.In
}

func NewRepository(params Params) *Repository {
	return &Repository{}
}
