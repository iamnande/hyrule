package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	svcfg "github.com/iamnande/hyrule/go/internal/lib/config"
)

type ServiceMetadata struct {
	// service
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`

	// runtime
	Environment svcfg.Environment `json:"environment"`
	Region      svcfg.Region      `json:"region"`

	// proof we're not stuck in time vortex
	Timestamp string `json:"timestamp"`

	// where to find more - source, docs, conventions. every service gets
	// this for free, human or agent hitting a running instance can always
	// find its own way home.
	Links map[string]string `json:"links,omitempty"`
}

func discoveryHandler(cfg *config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg.serviceMetadata.Timestamp = time.Now().UTC().Format(time.RFC3339)
		response, err := json.Marshal(cfg.serviceMetadata)
		if err != nil {
			cfg.logger.Error("failed to marshal service metadata", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err = w.Write(response); err != nil {
			cfg.logger.Error("failed to write service metadata", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
