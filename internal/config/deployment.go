package config

type Deployment struct {
	Region      Region      `env:"REGION" envDefault:"us-east-2"`
	Environment Environment `env:"ENVIRONMENT" envDefault:"local"`
}

func LoadDeployment() func() (Deployment, error) {
	return load[Deployment]("")
}
