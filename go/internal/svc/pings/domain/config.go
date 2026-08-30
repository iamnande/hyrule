package domain

import "time"

type Config struct {
	StaleAfter time.Duration `env:"STALE_AFTER" envDefault:"5m"`
}
