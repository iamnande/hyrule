package pings

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pings Suite")
}
