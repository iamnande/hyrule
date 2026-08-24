package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/iamnande/hyrule/internal/lib/tracing"
)

type DependencyStatusValue string

const (
	DependencyStatusUp   DependencyStatusValue = "up"
	DependencyStatusDown DependencyStatusValue = "down"
)

func (status DependencyStatusValue) String() string {
	return string(status)
}

type DependencyCheckFn func(context.Context) error

type dependencyCheck struct {
	name      string
	checkType DependencyCheckType
	check     DependencyCheckFn
}

type DependencyCheckType string

const (
	DependencyCheckTypeHard DependencyCheckType = "hard"
	DependencyCheckTypeSoft DependencyCheckType = "soft"
)

type DependencyStatus struct {
	Name   string                `json:"name"`
	Status DependencyStatusValue `json:"status"`
	Type   DependencyCheckType   `json:"type"`
}

type DependencyStatusResponse struct {
	Status       DependencyStatusValue `json:"status"`
	Dependencies []DependencyStatus    `json:"dependencies"`
}

func dependencyChecksHandler(cfg *config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err error

			dependencyCheckCount = int64(len(cfg.dependencyChecks))
			dependencyChecks     = sync.Map{}
			response             = DependencyStatusResponse{
				Status:       DependencyStatusUp,
				Dependencies: make([]DependencyStatus, len(cfg.dependencyChecks)),
			}
			responseController = http.NewResponseController(w)
			weightedSemaphore  = semaphore.NewWeighted(dependencyCheckCount)
		)
		spanCtx, done := tracing.StartNamed(r.Context(), "dependencies:check:all")
		defer done()
		tracing.SetTag(spanCtx, "dependencies.count", fmt.Sprintf("%d", len(cfg.dependencyChecks)))
		tracing.SetTag(spanCtx, "dependencies.timeout", cfg.timeout.String())

		// set the request write deadline to the configured value
		// we're also going to enforce this deadline with a context
		if err = responseController.SetWriteDeadline(time.Now().Add(cfg.timeout)); err != nil {
			cfg.logger.Error("failed to set response write deadline", slog.Any("error", err))
		}
		ctx, cancel := context.WithTimeout(spanCtx, cfg.timeout)
		defer cancel()

		// check all dependencies
		for _, check := range cfg.dependencyChecks {
			checkCtx, checkDone := tracing.StartNamed(ctx, fmt.Sprintf("dependencies:check:%s", check.name))
			if err = weightedSemaphore.Acquire(ctx, 1); err != nil {
				cfg.logger.Error("failed to acquire semaphore", slog.Any("error", err))
			}
			go func(ctx context.Context, statusTracker *sync.Map, dependency dependencyCheck) {
				defer checkDone()
				defer weightedSemaphore.Release(1)
				status := DependencyStatus{
					Name:   dependency.name,
					Status: DependencyStatusDown,
					Type:   dependency.checkType,
				}
				if err = dependency.check(ctx); err == nil {
					status.Status = DependencyStatusUp
				}
				dependencyChecks.Store(dependency.name, status)
				tracing.SetTag(ctx, fmt.Sprintf("dependencies:%s:status", dependency.name), status.Status.String())
			}(checkCtx, &dependencyChecks, check)
		}

		if err = weightedSemaphore.Acquire(ctx, dependencyCheckCount); err != nil {
			cfg.logger.Error(ErrDependencyRequestTimeout.Error(), slog.Any("error", err))
		}

		// build the response
		idx := 0
		dependencyChecks.Range(func(_, value any) bool {
			check, _ := value.(DependencyStatus)
			if check.Status == DependencyStatusDown && check.Type == DependencyCheckTypeHard {
				response.Status = DependencyStatusDown
			}
			response.Dependencies[idx] = check
			idx++
			return true
		})

		w.Header().Set("Content-Type", "application/json")
		if response.Status == DependencyStatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		tracing.SetTag(spanCtx, "dependencies:status", response.Status.String())
		body, err := json.Marshal(response)
		if err != nil {
			cfg.logger.Error("failed to marshal dependency check response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(body); err != nil {
			cfg.logger.Error("failed to write dependency check response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
