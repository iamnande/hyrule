package iamjwks

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIamJwks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "iam-jwks Suite")
}
