package config

type Password struct {
	Memory     uint32 `env:"MEMORY" envDefault:"65536"`
	Iterations uint32 `env:"ITERATIONS" envDefault:"3"`
	Threads    uint8  `env:"THREADS" envDefault:"4"`
	KeyLength  uint32 `env:"KEY_LENGTH" envDefault:"32"`
	SaltLength uint32 `env:"SALT_LENGTH" envDefault:"16"`

	// NOTE: This must be stored separately from the application.
	Pepper string `env:"PEPPER,required,unset"`
}

func LoadPassword() func() (Password, error) {
	return load[Password]("PASSWORD")
}
