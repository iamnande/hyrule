package config

type JWT struct {
	PrivateKey string `env:"PRIVATE_KEY,required,file" name:"service:jwt:private-key"`
	PublicKey  string `env:"PUBLIC_KEY,required,file" name:"service:jwt:public-key"`
}

func LoadJWT() func() (JWT, error) {
	return load[JWT]("JWT")
}
