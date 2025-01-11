package config

type Deployment struct {
	Region      Region      `env:"REGION,notEmpty"`
	Environment Environment `env:"ENVIRONMENT,notEmpty"`
}

func LoadDeployment(prefix string) func() (Deployment, error) {
	return load[Deployment](prefix)
}
