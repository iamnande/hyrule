package config

type Environment string

const (
	LocalEnvironment Environment = "local"
	DevEnvironment   Environment = "dev"
	ProdEnvironment  Environment = "prod"
)

var validEnvironments = []Environment{LocalEnvironment, DevEnvironment, ProdEnvironment}

func (e Environment) String() string {
	return string(e)
}
