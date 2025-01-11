package config

type Environment string

const (
	LocalEnvironment Environment = "local"
	DevEnvironment   Environment = "dev"
	ProdEnvironment  Environment = "prod"
)
