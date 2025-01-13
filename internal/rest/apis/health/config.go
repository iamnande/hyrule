package health

import (
	"log/slog"
	"time"
)

type config struct {
	logger           *slog.Logger
	timeout          time.Duration
	serviceMetadata  *ServiceMetadata
	dependencyChecks []dependencyCheck
}
