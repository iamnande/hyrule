package config

type Region string

const (
	USWest2Region Region = "us-west-2"
	USEast2Region Region = "us-east-2"
)

var validRegions = []Region{USWest2Region, USEast2Region}

func (r Region) String() string {
	return string(r)
}
