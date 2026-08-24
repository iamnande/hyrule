package config

import (
	"fmt"
	"net/url"
)

type Database struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	User     string `env:"USER" envDefault:"hyrule_app"`
	Password string `env:"PASSWORD" envDefault:"hyrule_app"`
	Name     string `env:"NAME" envDefault:"hyrule"`
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
	MaxConns int32  `env:"MAX_CONNS" envDefault:"10"`
}

func (db Database) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.User, db.Password),
		Host:     fmt.Sprintf("%s:%d", db.Host, db.Port),
		Path:     "/" + db.Name,
		RawQuery: "sslmode=" + db.SSLMode,
	}
	return u.String()
}

func LoadDatabase() func() (Database, error) {
	return load[Database]("DATABASE")
}
