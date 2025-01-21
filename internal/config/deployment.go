package config

type Deployment struct {
	Region      Region      `env:"REGION,notEmpty"`
	Environment Environment `env:"ENVIRONMENT,notEmpty"`
}

func LoadDeployment() func() (Deployment, error) {
	return load[Deployment]("")
}
