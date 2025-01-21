package config

type JWT struct {
	PrivateKey string `env:"PRIVATE_KEY,required"`
	PublicKey  string `env:"PUBLIC_KEY,required"`
}

func LoadJWT() func() (JWT, error) {
	return load[JWT]("JWT")
}
