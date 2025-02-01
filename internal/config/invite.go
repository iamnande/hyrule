package config

import (
	"time"
)

type Invite struct {
	ValidityWindow time.Duration `env:"VALIDITY_WINDOW" envDefault:"24h"`
}

func LoadInvite() func() (Invite, error) {
	return load[Invite]("INVITE")
}
