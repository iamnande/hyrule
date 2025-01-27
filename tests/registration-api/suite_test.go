package adminapi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAdminAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admin API Suite")
}
