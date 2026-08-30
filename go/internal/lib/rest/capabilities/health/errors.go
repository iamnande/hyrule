package health

import (
	"errors"
)

var (
	ErrStartupRequired          = errors.New("startup handler is required")
	ErrLivenessRequired         = errors.New("liveness handler is required")
	ErrReadinessRequired        = errors.New("readiness handler is required")
	ErrDependencyRequestTimeout = errors.New("dependency request exceeded the timeout")
)
