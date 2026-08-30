package config

type Environment string

const (
	LocalEnvironment Environment = "local"
	ProdEnvironment  Environment = "prod"
)

var validEnvironments = []Environment{LocalEnvironment, ProdEnvironment}

func (e Environment) String() string {
	return string(e)
}
