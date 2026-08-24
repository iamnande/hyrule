package tracing

import "github.com/getsentry/sentry-go"

type InitOptions struct {
	Enabled      bool
	SampleRate   float64
	IngestionURL string
	Release      string
	Environment  string
	Tags         map[string]string
}

// Init starts the tracing backend. A no-op if opts.Enabled is false.
func Init(opts InitOptions) error {
	if !opts.Enabled {
		return nil
	}
	return sentry.Init(sentry.ClientOptions{
		EnableTracing:    opts.Enabled,
		TracesSampleRate: opts.SampleRate,
		Dsn:              opts.IngestionURL,
		Release:          opts.Release,
		Environment:      opts.Environment,
		Tags:             opts.Tags,
	})
}
