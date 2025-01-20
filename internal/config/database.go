package config

type Database struct {
	Name          string `env:"NAME" envDefault:"hyrule"`
	LocalEndpoint string `env:"LOCAL_ENDPOINT" envDefault:"http://localhost:5432"`
}

func LoadDatabase(prefix string) func() (Database, error) {
	return load[Database](prefix)
}
