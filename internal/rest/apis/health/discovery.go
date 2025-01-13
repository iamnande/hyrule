package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type ServiceMetadata struct {
	// service
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`

	// runtime
	Environment string `json:"environment"`
	Region      string `json:"region"`

	// proof we're not stuck in time vortex
	Timestamp string `json:"timestamp"`
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
