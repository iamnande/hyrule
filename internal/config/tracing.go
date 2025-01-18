package config

type Tracing struct {
	Enabled      bool    `env:"ENABLED" envDefault:"true"`
	SampleRate   float64 `env:"SAMPLE_RATE" envDefault:"1.0"`
	IngestionURL string  `env:"INGESTION_URL"`
}

func LoadTracing(prefix string) func() (Tracing, error) {
	return load[Tracing](prefix)
}
