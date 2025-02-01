package config

type Email struct {
	From      string `env:"FROM" envDefault:"hyrule-noreply@morethq.com"`
	Host      string `env:"HOST" envDefault:"localhost"`
	Port      int    `env:"PORT" envDefault:"1025"`
	EnableTLS bool   `env:"TLS" envDefault:"false"`
	Username  string `env:"USERNAME"`
	Password  string `env:"PASSWORD,unset"`
}

func LoadEmail() func() (Email, error) {
	return load[Email]("EMAIL")
}
