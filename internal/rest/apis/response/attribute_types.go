package response

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func Email(email string) string {
	return strings.ToLower(email)
}

func Time(ts time.Time) string {
	return ts.Format(time.RFC3339)
}

func UUID(id uuid.UUID) string {
	return strings.ToLower(id.String())
}
