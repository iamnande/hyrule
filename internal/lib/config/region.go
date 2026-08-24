package config

type Region string

const (
	PrimaryRegion   Region = "us-east-2"
	SecondaryRegion Region = "eu-central-1"
)

var validRegions = []Region{PrimaryRegion, SecondaryRegion}

func (r Region) String() string {
	return string(r)
}
